package helper

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"os"
	"strings"
)

type EventInfo struct {
	EventName    string `json:"eventName"`
	DeploymentID string `json:"deploymentId"`
	UpdatedAt    string `json:"updatedAt"`
	Reason       string `json:"reason"`
}

func GetAwsDefaultRegion() string {
	val, exists := os.LookupEnv("AWS_REGION")
	if !exists || len(val) < 1 {
		val = os.Getenv("AWS_DEFAULT_REGION")
	}

	return val
}

func ReadAWSSecret(secretID string, awsSession *session.Session) (string, error) {
	*awsSession.Config.Region = GetAwsDefaultRegion()

	sessionSecretsManager := secretsmanager.New(awsSession)

	secretValue, err := sessionSecretsManager.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: &secretID})

	if err != nil {
		return "", fmt.Errorf("error getting secret from ID '%s': %w", secretID, err)
	}

	return *secretValue.SecretString, nil
}

func ReadAWSParameter(paramID string, awsSession *session.Session) (string, error) {
	*awsSession.Config.Region = GetAwsDefaultRegion()

	sessionAWSParameter := ssm.New(awsSession)

	param, err := sessionAWSParameter.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(paramID),
		WithDecryption: aws.Bool(false),
	})

	if err != nil {
		return "", fmt.Errorf("error getting param value from ID '%s': %w", paramID, err)
	}

	return *param.Parameter.Value, nil
}

func GetServiceNameFromARN(arnString string) (string, error) {
	arnSplit := strings.Split(arnString, "service/")

	if len(arnSplit) < 2 {
		return "", WrapError(fmt.Sprintf("Input '%s' did not have expected 'service/' prefix", arnString), nil)
	}

	return arnSplit[1], nil
}

func ParseEventDetails(request events.CloudWatchEvent) (EventInfo, error) {
	var eventInfo EventInfo

	err := json.Unmarshal(request.Detail, &eventInfo)

	if err != nil {
		return eventInfo, WrapError("Unexpected error unmarshaling event details", err)
	}

	if eventInfo.EventName == "" {
		return eventInfo, WrapError("'eventName' attribute not found on payload", nil)
	}
	if eventInfo.DeploymentID == "" {
		return eventInfo, WrapError("'deploymentId' attribute not found on payload", nil)
	}
	if eventInfo.UpdatedAt == "" {
		return eventInfo, WrapError("'updatedAt' attribute not found on payload", nil)
	}

	return eventInfo, nil
}
