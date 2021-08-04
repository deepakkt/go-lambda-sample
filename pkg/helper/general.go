package helper

import "strconv"

func GetDefaultHTTPTimeout() int {
	// putting this as a configurable parameter
	// return value in seconds

	defaultTimeout, err := strconv.Atoi(GetStringEnv("HTTP_TIMEOUT", "15"))

	if err != nil {
		defaultTimeout = 15
	}

	return defaultTimeout
}


func GetDeploymentUser() string {
	user := GetStringEnv("DEPLOYMENT_USER", "services@graphcms.com")
	return user
}
