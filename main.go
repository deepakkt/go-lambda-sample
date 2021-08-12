package main

import (
	"context"
	"deployment-notifications/pkg/helper"
	"deployment-notifications/pkg/validate"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

var awsSession *session.Session

type LambdaResponse struct {
	message string
}

func init() {
	runEnv, err := validate.EnvValidate()

	if err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}

	logRunEnv(runEnv)

	awsSession = session.Must(session.NewSession())
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

	if eventDetails.EventName != "SERVICE_DEPLOYMENT_COMPLETED" {
		msg := fmt.Sprintf("We received '%s' which we don't track. We only want 'SERVICE_DEPLOYMENT_COMPLETED'",
			eventDetails.EventName)
		return msg, errors.New(msg)
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
	log.Printf("SSM Slack Message Parameter Used: %s", runEnv["SSM_PARAMETER_MESSAGE_SLACK"])
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

	newRelicMapping, err := helper.ReadAWSParameter(runEnv["SSM_PARAMETER_NAME_NEW_RELIC"], awsSession)
	if err != nil {
		log.Printf("Error Reading SSM Parameter '%s': %v", runEnv["SSM_PARAMETER_NAME_NEW_RELIC"], err)
		return LambdaResponse{message: "SSM New Relic Parameter Read Failure"}, err
	}

	serviceNewRelicMap, err := helper.DecodeStringJSON(newRelicMapping)
	if err != nil {
		log.Printf("Error Decoding SSM Parameter '%s': %v", runEnv["SSM_PARAMETER_NAME_NEW_RELIC"], err)
		return LambdaResponse{message: "SSM New Relic Parameter Decode Failure"}, err
	}

	slackMapping, err := helper.ReadAWSParameter(runEnv["SSM_PARAMETER_NAME_SLACK"], awsSession)
	if err != nil {
		log.Printf("Error Reading SSM Parameter '%s': %v", runEnv["SSM_PARAMETER_NAME_SLACK"], err)
		return LambdaResponse{message: "SSM Slack Read Failure"}, err
	}

	serviceSlackMap, err := helper.DecodeSlackMapping(slackMapping)
	if err != nil {
		log.Printf("Error Decoding SSM Parameter '%s': %v", runEnv["SSM_PARAMETER_NAME_SLACK"], err)
		return LambdaResponse{message: "SSM Slack Parameter Decode Failure"}, err
	}

	defaultSlackWebhook := helper.GetDefaultWebhook(serviceSlackMap)

	if defaultSlackWebhook == "" {
		log.Printf("Webhook service for 'default-service' not defined in '%s'", runEnv["SSM_PARAMETER_NAME_SLACK"])
		return LambdaResponse{message: "Default Slack Webhook not defined"},
			errors.New("Default Slack Webhook not defined")
	}

	slackMessageTemplate, err := helper.ReadAWSParameter(runEnv["SSM_PARAMETER_MESSAGE_SLACK"], awsSession)
	if err != nil {
		log.Printf("Error Reading SSM Parameter '%s': %v", runEnv["SSM_PARAMETER_MESSAGE_SLACK"], err)
		return LambdaResponse{message: "SSM Slack Message Template Read Failure"}, err
	}

	ecsARN := request.Resources[0]
	ecsServiceName, err := helper.GetServiceNameFromARN(ecsARN)

	if err != nil {
		log.Printf("Error Parsing Service Name '%s': %v", ecsARN, err)
		return LambdaResponse{message: "ECS Service Name Parse Failure"}, err
	}

	newRelicTargetApp, ok := serviceNewRelicMap[ecsServiceName]

	if !ok {
		// this means that the mapping did not contain an entry
		// for the service which is notifying us - it either means
		// we missed to configure it or we don't care about this
		// but this Lambda has no choice but to exit
		log.Printf("We did not find a mapping for '%s'. Aborting notification", ecsARN)
		return LambdaResponse{message: "ECS Service not configured for notification"},
			errors.New("ECS Service Not Configured")
	}

	newRelicPayload := helper.GetNewRelicPayload(request)

	newRelicAPIToken, err := helper.ReadAWSSecret(runEnv["NEW_RELIC_API_TOKEN"], awsSession)
	if err != nil {
		log.Printf("Error Reading New Relic API Token Secret '%s': %v", runEnv["NEW_RELIC_API_TOKEN"], err)
		return LambdaResponse{message: "SSM New Relic Token Secret Read Failure"}, err
	}

	newRelicError := false
	slackError := false

	deployStatus, err := helper.PostNewRelicDeployment(newRelicPayload,
		runEnv["NEW_RELIC_BASE_DOMAIN"], newRelicTargetApp, newRelicAPIToken)

	if err != nil {
		newRelicError = true
		if deployStatus == 999 {
			log.Printf("New Relic submit aborted: %v", err)
		} else {
			log.Printf("New Relic submit failed with status: %d, %v", deployStatus, err)
		}
		log.Println("We will attempt slack notification")
	} else {
		log.Printf("New Relic Payload submitted: %v, status: %d", newRelicPayload, deployStatus)
	}

	slackPayload := helper.GenerateSlackNotificationStruct(request)
	slackStatus, err := helper.PostSlackMessage(slackMessageTemplate, slackPayload, defaultSlackWebhook)

	if err != nil {
		slackError = true
		log.Printf("Slack post failed with status: %d, %v", slackStatus, err)
		log.Println("We will attempt other webhooks, if available")
	}

	additionalWebhooks := helper.LocateValueMultiple(ecsServiceName, serviceSlackMap)

	if len(additionalWebhooks) > 0 {
		log.Printf("Additional webhooks defined for service '%s'", ecsServiceName)

		for _, webhook := range additionalWebhooks {
			slackStatus, err := helper.PostSlackMessage(slackMessageTemplate, slackPayload, webhook)

			if err != nil {
				slackError = true
				log.Printf("Slack post failed with status: %d, %v", slackStatus, err)
				log.Println("We will attempt other webhooks, if available")
			}
		}
	}

	if newRelicError {
		log.Println("New Relic submission did not complete")
	}

	if slackError {
		log.Println("Slack submission did not complete for one or more webhooks")
	}

	if !newRelicError && !slackError {
		return LambdaResponse{message: "Notification complete!"}, nil
	}

	return LambdaResponse{message: "Notification incomplete!"},
		helper.WrapError("One ore more notification failures", nil)
}

func main() {
	lambda.Start(HandleRequest)
}
