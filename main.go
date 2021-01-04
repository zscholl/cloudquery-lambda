package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cloudquery/cloudquery/cloudqueryclient"
)

var DRIVER string
var DSN string

type Request struct {
	TaskName string `json:"taskName"`
}

func LambdaHandler(ctx context.Context, req Request) (string, error) {
	return TaskExecutor(req.TaskName)
}

func TaskExecutor(taskName string) (string, error) {
	switch taskName {
	case "fetch":
		Fetch(DRIVER, DSN, false)
	case "policy":
		Policy(DRIVER, DSN, false)
	default:
		log.Printf("Unknown task: %s", taskName)
	}
	return fmt.Sprintf("Completed task %s", taskName), nil
}

// Fetches resources from a cloud provider and saves them in the configured database
func Fetch(driver, dsn string, verbose bool) {
	client, err := cloudqueryclient.New(driver, dsn, verbose)
	if err != nil {
		log.Fatalf("Unable to initialize client: %s", err)
	}
	err = client.Run("config.yml")
	if err != nil {
		log.Fatalf("Error fetching resources: %s", err)
	}
}

// Runs a policy SQL statement and returns results
func Policy(driver, dsn string, verbose bool) {
	fmt.Println("Running policy queries")
}
func main() {
	DRIVER = os.Getenv("CLOUDQUERY_DRIVER")
	DSN = os.Getenv("CLOUDQUERY_DATABASE_STRING")
	if env := os.Getenv("AWS_LAMBDA_RUNTIME_API"); env != "" {
		lambda.Start(LambdaHandler)
	} else if len(os.Args) > 1 {
		TaskExecutor(os.Args[1])
	} else {
		log.Fatalf("Usage: ./main [TASK]")
	}

}
