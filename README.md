# SSM Loader

## Input variables

|variable|example|required?|description|
|-|-|-|-|
|`SSM_PREFIX`|`/dev`|`false`|prefix for the SSM Parameters store path|
|`SSM_OUTPUT_DIR`|`/secrets`|`true`|output directory for secrets|
|`SSM_OUTPUT_FILENAME`|`secrets.json`|`true`|filename of the output configuration file, if `SSM_OUTPUT_DIR` is set to `/secrets`, the full path would be `/secrets/secrets.json`|
|`SSM_LOAD_`**`<variable name>`**|`/backend/random-api-key`|`false`|this variable contains path to the certain SSM Parameters store value, if `SSM_PREFIX` is set to `/dev`, the full path would be `/dev/backend/random-api-key`|
|`SSM_DEBUG`|`true`|`false`|if set to `true`, SSM loaders will log everything to the console (this might be dangerous)|

## AWS configuration

## Kubernetes example