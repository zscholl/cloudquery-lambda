package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	TaskName string `json:"taskName"`
}

func HandleRequest(ctx context.Context, req Request) (string, error) {
	return fmt.Sprintf("Running %s!", req.TaskName), nil
}

func main() {
	lambda.Start(HandleRequest)
}
