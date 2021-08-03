package main

import (
	"context"
	"deployment-notifications/pkg/helper"
	"deployment-notifications/pkg/validate"
	"github.com/aws/aws-lambda-go/lambda"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

type LambdaResponse struct {
	message string
}

func validators(request events.CloudWatchEvent) (string, error) {
	message, err := validate.SourceValidate(request)
	if err != nil {
		return message, err
	}

	message, err = validate.DetailValidate(request)

	if err != nil {
		return message, err
	}

	_, err = validate.ParseEventName(request)

	if err != nil {
		return "Event Name Parsing Error", err
	}

	_, err = validate.EnvValidate()

	if err != nil {
		return "Environment Validation Error", err
	}

	return "", nil
}


func HandleRequest(ctx context.Context, request events.CloudWatchEvent) (LambdaResponse, error) {

	errorMessage, err := validators(request)

	if err != nil {
		log.Printf("Error validating execution setup: %s", errorMessage)
		log.Printf("Error: %v", err)
		return LambdaResponse{message: errorMessage}, err
	}

	eventName, _ := validate.ParseEventName(request)
	runEnv, _ := validate.EnvValidate()
	ecsARN := request.Resources[0]

	log.Printf("Event Source: %s", request.Source)
	log.Printf("Event ID: %s", request.ID)
	log.Printf("Event Detail Type: %s", request.DetailType)
	log.Printf("Event Region: %s", request.Region)
	log.Printf("Event Timestamp: %s", request.Time)
	log.Printf("Event Name: %s", eventName)
	log.Printf("ECS ARN: %s", ecsARN)

	log.Printf("SSM Parameter Used: %s", runEnv["SSM_PARAMETER_NAME"])

	paramValue, _ := helper.ReadAWSParameter(runEnv["SSM_PARAMETER_NAME"])
	log.Printf("Parameter 'SSM_PARAMETER_NAME' value is '%s'", paramValue)

	return LambdaResponse{message: "Notification complete!"}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
