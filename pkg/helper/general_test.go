package helper_test

import (
	"deployment-notifications/pkg/helper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)


func TestHTTPTimeout(t *testing.T) {
	defaultTime := helper.GetDefaultHTTPTimeout()
	assert.Equal(t, 15, defaultTime)

	os.Setenv("HTTP_TIMEOUT", "999")
	defer os.Unsetenv("HTTP_TIMEOUT")
	defaultTime = helper.GetDefaultHTTPTimeout()
	assert.Equal(t, 999, defaultTime)
}


func TestDefaultUser(t *testing.T) {
	defaultUser := helper.GetDeploymentUser()
	assert.Equal(t, "services@graphcms.com", defaultUser)

	os.Setenv("DEPLOYMENT_USER", "whoami@graphcms.com")
	defer os.Unsetenv("DEPLOYMENT_USER")
	defaultUser = helper.GetDeploymentUser()
	assert.Equal(t, "whoami@graphcms.com", defaultUser)
}