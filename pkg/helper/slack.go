package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type SlackNotificationFields struct {
	ServiceName string
	DeploymentRevision string
	AWSReference string
	AWSRegion string
	AWSAccount string
	DeploymentTimestamp string
	DeploymentDescription string
}


func DecodeSlackMapping(parameterString string) (map[string][]string, error) {
	//this function assumes that the parameter string is a series of
	//key value pairs which are all string. Any other input type will
	//error out and the error is returned as is

	resultMap := make(map[string][]string)
	err := json.Unmarshal([]byte(parameterString), &resultMap)

	if err != nil {
		return resultMap, WrapError(fmt.Sprintf("Could not decode slack parameter string\n%s", parameterString),
			nil)
	}

	return resultMap, nil
}


func GetDefaultWebhook(valueMap map[string][]string) string {
	defaultServiceMap := LocateValueMultiple("default-service", valueMap)

	if len(defaultServiceMap) == 0 {
		return ""
	}

	return defaultServiceMap[0]
}


func GenerateSlackNotificationStruct(request events.CloudWatchEvent) SlackNotificationFields {
	eventDetails, _ := ParseEventDetails(request)
	ecsServiceName, _ := GetServiceNameFromARN(request.Resources[0])

	return SlackNotificationFields{
		ServiceName: ecsServiceName,
		DeploymentRevision: eventDetails.DeploymentID,
		AWSReference: request.ID,
		AWSRegion: request.Region,
		AWSAccount: request.AccountID,
		DeploymentTimestamp: eventDetails.UpdatedAt,
		DeploymentDescription: eventDetails.Reason,
	}
}


func GeneratePayload(templateMessage string, templateValues SlackNotificationFields, parseQuoteTags bool) (string, error) {
	parsedMessage := templateMessage

	parsedMessage = strings.ReplaceAll(templateMessage, "<varbegin>", "{{")
	parsedMessage = strings.ReplaceAll(parsedMessage, "<varend>", "}}")

	t := template.New("Slack Template")

	t, err := t.Parse(parsedMessage)

	if err != nil {
		return "", WrapError("error parsing Slack message template", err)
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, templateValues)

	if err != nil {
		return "", WrapError("error applying values to Slack message template", err)
	}

	finalOut := tpl.String()

	if parseQuoteTags {
		finalOut = strings.ReplaceAll(finalOut, "<backquote>", "`")
	}

	return finalOut, nil
}


func PostSlackMessage(messageTemplate string, templateValues SlackNotificationFields,
	webhookURL string) (int, error) {

	parsedMessage, err := GeneratePayload(messageTemplate, templateValues, true)

	fmt.Println(parsedMessage)

	if err != nil {
		return 500, WrapError("Error parsing slack message template", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer([]byte(parsedMessage)))

	if err != nil {
		return 500, WrapError("Error formatting new request for New Relic deployment", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(GetDefaultHTTPTimeout())

	resp, err := client.Do(req)
	if err != nil {
		return resp.StatusCode, WrapError("Error making final Slack request", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %d", resp.StatusCode)
	respBody, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Response Body: %s", string(respBody))

	return resp.StatusCode, nil
}