package main

import (
	"context"
	"deployment-notifications/pkg/helper"
	"deployment-notifications/pkg/validate"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
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

	eventDetails, err := helper.ParseEventDetails(request)

	if err != nil {
		return "Event Details Parsing Error", err
	}

	if eventDetails["eventName"] != "SERVICE_DEPLOYMENT_COMPLETED" {
		msg := fmt.Sprintf("We received '%s' which we don't track. We only want 'SERVICE_DEPLOYMENT_COMPLETED'",
			eventDetails["eventName"])
		return msg, errors.New(msg)
	}

	_, err = validate.EnvValidate()

	if err != nil {
		return "Environment Validation Error", err
	}

	return "", nil
}

func logRequest(request events.CloudWatchEvent) {
	log.Printf("Event Source: %s", request.Source)
	log.Printf("Event ID: %s", request.ID)
	log.Printf("Event Detail Type: %s", request.DetailType)
	log.Printf("Event Region: %s", request.Region)
	log.Printf("Event Timestamp: %s", request.Time)
	log.Printf("Event Name: 'SERVICE_DEPLOYMENT_COMPLETED'")
	log.Printf("ECS ARN: %s", request.Resources[0])
}

func logRunEnv(runEnv map[string]string) {
	log.Printf("SSM New Relic Parameter Used: %s", runEnv["SSM_PARAMETER_NAME_NEW_RELIC"])
	log.Printf("SSM Slack Parameter Used: %s", runEnv["SSM_PARAMETER_NAME_SLACK"])
	log.Printf("New Relic API Token Secret Name: %s", runEnv["NEW_RELIC_API_TOKEN"])
	log.Printf("Slack Token Secret Name: %s", runEnv["SLACK_API_TOKEN"])
	log.Printf("New Relic Base Domain for API Calls: %s", runEnv["NEW_RELIC_BASE_DOMAIN"])
	log.Printf("AWS Account Number: %s", runEnv["AWS_ACCOUNT_NUMBER"])
	log.Printf("AWS Region: %s", helper.GetAwsDefaultRegion())
}


func HandleRequest(ctx context.Context, request events.CloudWatchEvent) (LambdaResponse, error) {
	errorMessage, err := validators(request)

	if err != nil {
		log.Printf("Error validating execution setup: %s", errorMessage)
		log.Printf("Error: %v", err)
		return LambdaResponse{message: errorMessage}, err
	}

	logRequest(request)
	runEnv, _ := validate.EnvValidate()
	logRunEnv(runEnv)

	newRelicMapping, err := helper.ReadAWSParameter(runEnv["SSM_PARAMETER_NAME_NEW_RELIC"])
	if err != nil {
		log.Printf("Error Reading SSM Parameter '%s': %v", runEnv["SSM_PARAMETER_NAME_NEW_RELIC"], err)
		return LambdaResponse{message: "SSM New Relic Parameter Read Failure"}, err
	}

	serviceNewRelicMap, err := helper.DecodeStringJSON(newRelicMapping)
	if err != nil {
		log.Printf("Error Decoding SSM Parameter '%s': %v", runEnv["SSM_PARAMETER_NAME_NEW_RELIC"], err)
		return LambdaResponse{message: "SSM New Relic Parameter Decode Failure"}, err
	}

	ecsARN := request.Resources[0]
	ecsServiceName, err := helper.GetServiceNameFromARN(ecsARN)

	if err != nil {
		log.Printf("Error Parsing Service Name '%s': %v", ecsARN, err)
		return LambdaResponse{message: "ECS Service Name Parse Failure"}, err
	}

	newRelicTargetApp := helper.LocateValue(ecsServiceName, serviceNewRelicMap)

	if newRelicTargetApp == "" {
		// this means that the mapping did not contain an entry
		// for the service which is notifying us - it either means
		// we missed to configure it or we don't care about this
		// but this Lambda has no choice but to exit
		log.Printf("We did not find a mapping for '%s'. Aborting notification", ecsARN)
		return LambdaResponse{message: "ECS Service not configured for notification"},
			errors.New("ECS Service Not Configured")
	}

	newRelicPayload := helper.GetNewRelicPayload(request)

	newRelicAPIToken, err := helper.ReadAWSSecret(runEnv["NEW_RELIC_API_TOKEN"])
	if err != nil {
		log.Printf("Error Reading New Relic API Token Secret '%s': %v", runEnv["NEW_RELIC_API_TOKEN"], err)
		return LambdaResponse{message: "SSM New Relic Token Secret Read Failure"}, err
	}

	deployStatus, err := helper.PostDeploymentPayload(newRelicPayload,
		runEnv["NEW_RELIC_BASE_DOMAIN"], newRelicTargetApp, newRelicAPIToken)
	if err != nil {
		log.Printf("New Relic submit failed with status: %d, %v", deployStatus, err)
		return LambdaResponse{message: "New Relic submission failure"}, err
	}

	log.Printf("New Relic Payload submitted: %v, status: %d", newRelicPayload, deployStatus)
	return LambdaResponse{message: "Notification complete!"}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
