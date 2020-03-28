package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"log"
	"os"
	"strconv"
	"strings"
)

// Output is used only for the formatting of JSON
// payload
type Output struct {
	Parameters []*Param `json:"parameters"`
}

// GlobalConfiguration contains program configuration
// except the list of parameters
type GlobalConfiguration struct {
	OutputDirectory string
	OutputFilename  string
	Debug           bool
}

// Param holds input specs and output value
// for the further processing
type Param struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Path    string `json:"path"`
	Version int64  `json:"version"`
	Digest  string `json:"digest"`
}

// for debug only!
func (i *Param) String() string {
	return fmt.Sprintf("name = %s Value = %s Path = %s\n", i.Name, i.Value, i.Path)
}

func main() {
	log.Println("starting execution")

	// get global configuration
	log.Println("loading global configuration")
	config, err := loadGlobalConfiguration()
	if err != nil {
		log.Fatalf("something went wrong while loading global configuration: %s", err)
	}

	// get list of variables
	params := getParametersSpecification()
	if cap(params) == 0 {
		log.Println("no parameters provided, exiting")
		return
	}

	// initialize AWS API session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		// tbd: add role ARN
	}))

	// create SSM service
	ssmSvc := ssm.New(sess)

	// get values from AWS SSM
	err = obtainParams(ssmSvc, params)
	if err != nil {
		log.Fatalf("could not obtain params: %s", err)
	}
	if config.Debug {
		log.Printf("Obtained parameters: %s", params)
	}

	// create configuration file
	err = dumpToFile(params, config)
	if err != nil {
		log.Fatalf("could not create output file: %s", err)
	}

	return
}

// loadGlobalConfiguration load global configuration
func loadGlobalConfiguration() (GlobalConfiguration, error) {
	config := GlobalConfiguration{}
	config.OutputDirectory = os.Getenv("SSM_OUTPUT_DIR")
	config.OutputFilename = os.Getenv("SSM_OUTPUT_FILENAME")

	if config.OutputDirectory == "" || config.OutputFilename == "" {
		return config, fmt.Errorf("SSM_OUTPUT_DIR and SSM_OUTPUT_FILENAME must be provided")
	}

	// try to parse debug
	debug, err := strconv.ParseBool(os.Getenv("SSM_DEBUG"))
	if err != nil {
		debug = false
	}
	config.Debug = debug

	return config, nil
}

// dumpToFile create the configuration file that can be loaded by
// the application
func dumpToFile(payload []*Param, config GlobalConfiguration) error {

	// create output in the expected format
	output := Output{
		Parameters: payload,
	}

	// serialize output to JSON
	jsonPayload, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("could not marshal payload to JSON: %s", err)
	}

	// show JSON string if needed
	if config.Debug {
		log.Printf("json payload: %s", jsonPayload)
	}

	// assemble full path of the output file, additional slash is ok
	fullPath := config.OutputDirectory + "/" + config.OutputFilename

	// create config file to the requested location
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("could not create output file: %s", err)
	}
	defer f.Close()
	log.Printf("writing configuration file to %s", fullPath)

	// write JSON payload to the requested location
	_, err = f.WriteString(string(jsonPayload))
	if err != nil {
		return fmt.Errorf("could not write JSON payload to file: %s", err)
	}

	// all done, cool
	return nil
}

// obtainParams receives specification of needed values and mutates
// the input with received values (and versions)
func obtainParams(ssmSvc ssmiface.SSMAPI, params []*Param) error {
	for i := range params {
		log.Printf("obtaining param '%s' @ %s", params[i].Name, params[i].Path)

		// prepare AWS input struct
		getParamInput := ssm.GetParameterInput{
			Name:           aws.String(params[i].Path),
			WithDecryption: aws.Bool(true),
		}

		// try to obtain param from SSM Parameters Store,
		// fail if any error occurs
		paramOutput, err := ssmSvc.GetParameter(&getParamInput)
		if err != nil {
			return fmt.Errorf("could not obtain param '%s': %s", params[i].Name, err)
		}

		// mutate Value in the original specification slice
		params[i].Value = *paramOutput.Parameter.Value
		params[i].Version = *paramOutput.Parameter.Version
	}
	return nil
}

// getParametersSpecification is getting environment variables and
// converts them to more convenient format: slice of Param
func getParametersSpecification() (output []*Param) {
	// get all prefix for parameter paths e.g. /dev1, /production etc.
	prefix := os.Getenv("SSM_PREFIX")

	// convert environment variables to slice of Param
	for _, e := range os.Environ() {
		envPair := strings.SplitN(e, "=", 2)
		if strings.Contains(envPair[0], "SSM_LOAD_") {

			// get name without SSM_LOAD_ i.e. SSM_LOAD_test becomes test
			name := strings.Replace(envPair[0], "SSM_LOAD_", "", -1)

			// append current param to output slice
			param := Param{
				Name: name,
				Path: prefix + envPair[1],
			}
			output = append(output, &param)
		}
	}
	return
}
