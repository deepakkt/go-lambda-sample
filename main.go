package main

import (
	"context"
	"deployment-notifications/pkg/validate"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type LambdaResponse struct {
	message string
}


func HandleRequest(ctx context.Context, request events.CloudWatchEvent) (LambdaResponse, error) {
	message, err := validate.SourceValidate(request)

	if err != nil {
		log.Println(message)
		return LambdaResponse{message: message}, err
	}

	message, err = validate.DetailValidate(request)

	if err != nil {
		log.Println(message)
		return LambdaResponse{message: message}, err
	}

	eventName, err := validate.ParseEventName(request)

	if err != nil {
		log.Printf("Event Name Parsing Error: %v", err)
		return LambdaResponse{message: "Event Name Parsing Error"}, err
	}

	ecsARN := request.Resources[0]

	log.Printf("Event Source: %s", request.Source)
	log.Printf("Event ID: %s", request.ID)
	log.Printf("Event Detail Type: %s", request.DetailType)
	log.Printf("Event Region: %s", request.Region)
	log.Printf("Event Timestamp: %s", request.Time)
	log.Printf("Event Name: %s", eventName)
	log.Printf("ECS ARN: %s", ecsARN)

	return LambdaResponse{message: "Notification complete!"}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
