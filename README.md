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
