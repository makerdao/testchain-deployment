# testchain-deployment

## Requirements

* go > 1.11
* enabled `go mod` for pck versioning
* docker for run
* github.com api token for run updating of deployment source(simple put it to file `.apiToken` in the root of project)

## Build and run

### Local

* install dapp and all requirements for [deployment scripts](https://github.com/makerdao/testchain-dss-deployment-scripts)
* Run application
  * `TCD_DEPLOY="deploymentDirPath=$HOME/deployment" make run GOOS=darwin` - for mac
  * `TCD_DEPLOY="deploymentDirPath=$HOME/deployment" make run GOOS=linux` - for linux

### Docker(prefer)

* run `make build-base-image` - only first time, create big image with dapp, bash, git.
* use `make run-image` - for rebuild rerun app
* use `make logs` - for show logs -f from container

## Configuring

U can see all config in `pkg/config/config.go`

Default config is ready for local docker

For using custom config var u must use ENV variables.

## API

Protocol based on json object in http body.
Supply only `POST` requests.

Use id for operation, if u run long time operation, u get response ok on request ASAP 
and after async running service will send to gateway result of operation. 

### Models

Request example: 
```json
{
  "id": "uniqID_for_operation",
  "method": "method_name",
  "data": {"someKey": "someVal"} 
}
```

Response example:
```json
{
  "type": "ok || error",
  "result": {"someKey": "someVal"}
}
```

Error object example:
```json
{
  "code": "code",
  "detail": "detail error",
  "errorList": ["error text1", "error text 2"]
}
```

List of codes:
* internalError
* badRequest
* notFound

### Methods:

#### GetInfo

Request:
```json
{
  "id": "reqID",
  "method": "GetInfo",
  "data": {}
}
```

Good response example:
```json
{
  "type": "ok",
  "result": {
    "updatedAt": "2019-01-27T13:07:09.173377348Z", // last update dt
    "steps": [ // list of available step with full info
      {
        "id": 1,
        "description": "Step 7 - MS 3 - Crash & Bite",
        "defaults": {
          "osmDelay": "1"
        },
        "roles": [
          "CREATOR",
          "CDP_OWNER"
        ],
        "oracles": [
          {
            "symbol": "ETH",
            "contract": "MEDIANIZER_ETH"
          },
          {
            "symbol": "REP",
            "contract": "MEDIANIZER_REP"
          }
        ]
      }
    ],
    "tagHash": "f1e23cd2aecb42ddb74f29eb7db576f21b1911d9\n" // hash of commit for tag
  }
}
```

#### Run

Request:
```json
{
  "id": "reqID",
  "method": "Run",
  "data": {
    "stepId": 1, // number of step
    "envVars": { // map of env vars for run cmd
      "NAME_OF_ENV_VAR": "valueOfEnvVar"
    }
  }
}
```

Good response example:
```json
{
  "type": "ok",
  "result": {}
}
```
_*When run is finished, system will send result to gateway_

#### UpdateSource

Request:
```json
{
  "id": "reqID",
  "method": "UpdateSource",
  "data": {}
}
```

Good response example:
```json
{
  "type": "ok",
  "result": {}
}
```
_*When update is finished, system will send result to gateway_
