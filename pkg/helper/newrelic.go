package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func GetNewRelicDeploymentURL(baseDomain, appID string) string {
	return fmt.Sprintf("https://%s/v2/applications/%s/deployments.json",
		baseDomain, appID)
}


func GetNewRelicPayload(request events.CloudWatchEvent) map[string]string {
	eventDetails, _ := ParseEventDetails(request)
	result := make(map[string]string)

	result["revision"] = eventDetails.DeploymentID
	result["timestamp"] = eventDetails.UpdatedAt
	result["user"] = GetDeploymentUser()
	result["description"] = fmt.Sprintf("AWS Account: %s, Region: %s, Deployment ID: %s",
		request.AccountID, request.Region, request.ID)
	result["changelog"] = eventDetails.Reason

	return result
}


func PostNewRelicDeployment(payload map[string]string,
	baseDomain, appID, apiKey string) (int, error) {
	// posts deployment payload to the New Relic application deployment
	// section. It adds the "deployment" meta-key
	// Do not include that in the input payload

	deploymentURL := GetNewRelicDeploymentURL(baseDomain, appID)

	finalPayload := make(map[string]map[string]string)
	finalPayload["deployment"] = payload
	finalPayloadBytes, err := json.Marshal(finalPayload)

	if err != nil {
		return 999, WrapError("Error marshaling New Relic deployment payload into bytes", err)
	}

	req, err := http.NewRequest("POST", deploymentURL, bytes.NewBuffer(finalPayloadBytes))

	if err != nil {
		return 999, WrapError("Error formatting new request for New Relic deployment", err)
	}

	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	client.Timeout = time.Second * time.Duration(GetDefaultHTTPTimeout())

	resp, err := client.Do(req)
	if err != nil {
		return resp.StatusCode, WrapError("Error making final New Relic request", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status: %d", resp.StatusCode)
	respBody, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Response Body: %s", string(respBody))

	if resp.StatusCode != 201 {
		return resp.StatusCode, WrapError("New Relic final submission failed", nil)
	}
	return resp.StatusCode, nil
}
