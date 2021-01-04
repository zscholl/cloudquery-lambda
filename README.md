## ☁️ [Cloudquery](https://github.com/cloudquery/cloudquery) in AWS Lambda
This repo contains a reference implementation for running [Cloudquery](https://github.com/cloudquery/cloudquery) in AWS lambda.

## Usage
Install [Task](https://taskfile.dev/#/installation)

Build docker image locally:

`task docker:build`

Build docker test image with AWS Lambda runtime environment support

`task test:docker:build`

Run docker image locally:

`task test:docker:run`

`http post http://localhost:8080/2015-03-31/functions/function/invocations taskName=fetch`


## Deploy
TODO

