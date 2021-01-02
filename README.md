## ☁️ [Cloudquery](https://github.com/cloudquery/cloudquery) lambda reference implementation
This repo contains a reference implementation for running [Cloudquery](https://github.com/cloudquery/cloudquery) in AWS lambda.

## Usage
Install [Task](https://taskfile.dev/#/installation)

Build docker image locally:
`task docker:build`

Run docker image locally:
`task docker:run`

`http post http://localhost:8080/2015-03-31/functions/function/invocations taskName=test`

