package helper_test

import (
	"deployment-notifications/pkg/helper"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetNewRelicDeploymentURL(t *testing.T) {
	deploymentURL := helper.GetNewRelicDeploymentURL("onenewrelic.com", "12345")

	assert.Equal(t, "https://onenewrelic.com/v2/applications/12345/deployments.json", deploymentURL)
}

func TestGetNewRelicPayload(t *testing.T) {
	sampleEvent := `
{                                                                         
   "version": "0",                                                        
   "id": "ddca6449-b258-46c0-8653-e0e3a6EXAMPLE",                         
   "detail-type": "ECS Deployment State Change",                          
   "source": "AWS.ECS",                                                   
   "account": "111122223333",                                             
   "time": "2020-05-23T12:31:14Z",                                        
   "region": "us-west-2",                                                 
   "resources": [                                                         
        "arn:aws:ecs:us-west-2:111122223333:service/shure-content-api"    
   ],                                                                     
   "detail": {                                                            
        "eventType": "INFO",                                              
        "eventName": "SERVICE_DEPLOYMENT_COMPLETED",                      
        "deploymentId": "ecs-svc/123",                                    
        "updatedAt": "2020-05-23T11:11:11Z",                              
        "reason": "ECS deployment deploymentId in progress."              
   }                                                                      
}                                                                         
`
	currentDeploymentUser := os.Getenv("DEPLOYMENT_USER")
	os.Unsetenv("DEPLOYMENT_USER")
	defer os.Setenv("DEPLOYMENT_USER", currentDeploymentUser)

	var cloudwatchEvent events.CloudWatchEvent
	err := json.Unmarshal([]byte(sampleEvent), &cloudwatchEvent)
	assert.Nil(t, err)

	newRelicMap := helper.GetNewRelicPayload(cloudwatchEvent)

	assert.Equal(t, "ecs-svc/123", newRelicMap["revision"])
	assert.Equal(t, "2020-05-23T11:11:11Z", newRelicMap["timestamp"])
	assert.Equal(t, "services@graphcms.com", newRelicMap["user"])
	assert.Equal(t, "ECS deployment deploymentId in progress.", newRelicMap["changelog"])
	assert.Equal(t,
		"AWS Account: 111122223333, Region: us-west-2, Deployment ID: ddca6449-b258-46c0-8653-e0e3a6EXAMPLE",
		newRelicMap["description"])
}
