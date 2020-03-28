package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/aws/session"
	"strings"
	"fmt"
	"log"
	"os"
)

type Param struct {
	Name string
	Value string
	Path string
	Version int64
}

func main() {

	// get list of variables
	params := getParametersSpecification()
	
	// initialize AWS API session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		// tbd: add role ARN
	}))
	
	// create SSM service
	ssmSvc := ssm.New(sess)

	for i := range params {
		log.Printf("obtaining param '%s' @ %s", params[i].Name, params[i].Path)

		// prepare AWS input struct
		getParamInput := ssm.GetParameterInput{
			Name: aws.String(params[i].Path),
			WithDecryption: aws.Bool(true),
		}

		// try to obtain param from SSM Parameters Store,
		// fail if any error occurs
		paramOutput, err := ssmSvc.GetParameter(&getParamInput)
		if err != nil {
			log.Fatalf("could not obtain param: %s", err)
		}

		// mutate Value in the original specification slice
		params[i].Value = *paramOutput.Parameter.Value
		params[i].Version = *paramOutput.Parameter.Version
	}

	fmt.Println(params)
}

// getParametersSpecification is getting environment variables and
// converts them to more convenient format: slice of Param
func getParametersSpecification() (output []Param) {
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
			output = append(output, param)
		}
	}
	return
}