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

`GITHUB_DEFAULT_CHECKOUT_TARGET` - u can set default target for github chekout
(default: 'tags/qa-deploy')

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

#### GetResult

Request:
```json
{
  "id": "reqID",
  "method": "GetResult",
  "data": {}
}
```

Good response example:
```json
{
  "type": "ok",
  "result": {
    "lastUpdated": "2019-01-31T16:41:12.380468999Z",
    "data": {
      "MCD_DEPLOY": "0xc0786dc1d40be13a79b904eb6e35f2520ca6af05",
      "MCD_GOV": "0xa3f164797c8541b9ea2c86fa65c410cf80c1e40e",
      "MCD_GOV_GUARD": "0xdb4f343876ee32f050be0222f7c4c77bab9e73dd",
      "MCD_ADM": "0xd7ef0b16fcdd52bd40558809b21582b88efcd8b5",
      "MCD_VAT": "0x8281e7d6f955d0cd9af747525cbe0d985af241e7",
      "MCD_PIT": "0xa9938c15462e61983a472a724aab7e0d872f93ea",
      "MCD_DRIP": "0x7023c1bedacfc36985be7632d6ea8fe1d5fe4376",
      "MCD_CAT": "0xccb8880f91b2a7c354b90f7af2eef01cabab8ad3",
      "MCD_VOW": "0xe1b90bb70af38420f257fc1a10676c0d13efb57b",
      "MCD_JOIN_DAI": "0x6be8bd7d5585e016635fa6fb32c3da497d9a38d7",
      "MCD_MOVE_DAI": "0x2cc889ac3251472a21547efca262bc4fad102b97",
      "MCD_FLAP": "0x23d94e77d90d4f2262cdbad7e15144dac42f5b07",
      "MCD_FLOP": "0xe1b34872d7f7acc92f18c447affaadf66c54dee6",
      "MCD_MOM": "0xb3e682df241e702e38f702cf77e870a190711446",
      "MCD_MOM_LIB": "0x072b4d8619512119a79651f171a981ee2655cf8a",
      "MCD_DAI": "0xed307ee64a6f77da04e4a22ddbd49bb114da79fc",
      "MCD_DAI_GUARD": "0x0f0ac984e5d72e57abc8e1d644b9b8c35d9b8266",
      "MCD_SPOT": "0x74d58189177ca7829176a40a427a5148b1d36972",
      "MCD_POT": "0xb2d929077066c022cd041eff33c8f0aade39ef11",
      "PROXY_ACTIONS": "0x084eb4c7e045705e57041e5da23ab3ca21c056da",
      "CDP_MANAGER": "0x485341fb88327b22db31a1d4dbb66d00e09ad89e",
      "PROXY_FACTORY": "0x34a7224139619e6921743d316ce29dae9f98b98f",
      "PROXY_REGISTRY": "0x5b763736b398570044aa6c6c25d25abc54ce5a55",
      "PIP_ETH": "0x61c3488d54bb965f6a8aeabb0f5010a3aa5a51af",
      "VAL_ETH": "0x61c3488d54bb965f6a8aeabb0f5010a3aa5a51af",
      "MCD_JOIN_ETH": "0x12db7ed4071952cdd6107f5266c17dc658483f0d",
      "MCD_MOVE_ETH": "0xf61b9e634cd866b01511e5a3bfb19b782cd74238",
      "MCD_FLIP_ETH": "0xe7472b343389b82ec07514dad0adf94b72259474",
      "REP": "0x893ef8b7b809e14e87820e8ecf23950b0848cae4",
      "VAL_REP": "0xd3a53f0ca30e3511fa3b443070f712afc65a754f",
      "PIP_REP": "0xd3a53f0ca30e3511fa3b443070f712afc65a754f",
      "MCD_JOIN_REP": "0xbda89568548030f8643e8b3db1d8a936f6f10395",
      "MCD_MOVE_REP": "0xa59a7c4b2ce773242e89b673db6ce15f694fd6b6",
      "MCD_FLIP_REP": "0x86759f871e3e74d46e320d97a2554106e6191c51"
    }
  }
}
```

#### Checkout

Request:
```json
{
  "id": "reqID",
  "method": "Checkout",
  "data": {
    "commit": "hash_commit"
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

_*When update is finished, system will send result to gateway_

#### Checkout

Request:
```json
{
  "id": "reqID",
  "method": "GetCommitList",
  "data": {}
}
```

Good response example:
```json
{
  "type": "ok",
  "result": {
    "data": [
      {
        "commit" : "hash_commit",
        "author" : "name <email>",
        "date" : "readable date",
        "text" : "text of commit"
      },
      {
        "commit" : "hash_commit",
        "author" : "name <email>",
        "date" : "readable date",
        "text" : "text of commit"
      }
    ]
  }
}
```

## Examples

You can see example of http request in `./examples/http`.

## NATS.io

Supported async result for `Run` and `UpdateSource`.

### Example

1. `dc up`
2. `telnet localhost 4222` and insert `SUB Prefix.RunResult.* id1`
3. Send http request for run, for example `./examples/http/Run.http`
4. ...
5. PROFIT!!!!1111



