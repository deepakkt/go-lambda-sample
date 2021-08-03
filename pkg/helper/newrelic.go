package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetNewRelicDeploymentURL(baseDomain, appID string) string {
	return fmt.Sprintf("https://%s/v2/applications/%s/deployments.json",
		baseDomain, appID)
}


func GetDefaultHTTPTimeout() int {
	// putting this as a configurable parameter
	// return value in seconds

	defaultTimeout, err := strconv.Atoi(GetStringEnv("HTTP_TIMEOUT", "15"))

	if err != nil {
		defaultTimeout = 15
	}

	return defaultTimeout
}


func PostDeploymentPayload(payload map[string]string,
	baseDomain, appID, apiKey string) (statusCode int, err error) {
	// posts deployment payload to the New Relic application deployment
	// section. It adds the "deployment" meta-key
	// Do not include that in the input payload

	deploymentURL := GetNewRelicDeploymentURL(baseDomain, appID)

	finalPayload := make(map[string]map[string]string)
	finalPayload["deployment"] = payload
	finalPayloadBytes, err := json.Marshal(finalPayload)

	if err != nil {
		return 500, WrapError("Error marshaling New Relic deployment payload into bytes", err)
	}

	req, err := http.NewRequest("POST", deploymentURL, bytes.NewBuffer(finalPayloadBytes))

	if err != nil {
		return 500, WrapError("Error formatting new request for New Relic deployment", err)
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

	return resp.StatusCode, nil
}
