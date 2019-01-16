# testchain-deployment

## Requirements

* go > 1.11
* enabled `go mod` for pck versioning
* docker for run
* github.com api token for run updating of deployment source(simple put it to file `.apiToken` in the root of project)

## Build and run

See `Makefile`

## Configuring

U can see all config in `pkg/config/config.go`

Default config is ready for local docker

For using custom config var u must use ENV variables.
