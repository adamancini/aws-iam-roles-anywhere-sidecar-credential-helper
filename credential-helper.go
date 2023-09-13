package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	awsconfig "github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper/awsconfig"
)

func main() {

	credentialsURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")
	if credentialsURI == "" {
		fmt.Println("AWS_CONTAINER_CREDENTIALS_FULL_URI environment variable not set. defaulting to localhost:8080/creds")
		credentialsURI = "http://localhost:8080/creds"
	}

	refreshIntervalStr := os.Getenv("AWS_REFRESH_INTERVAL")
	refreshInterval := 300 // Default to 5 minutes (300 seconds)
	if refreshIntervalStr != "" {
		interval, err := strconv.Atoi(refreshIntervalStr)
		if err != nil {
			fmt.Println("Invalid AWS_REFRESH_INTERVAL:", err)
			return
		}
		refreshInterval = interval
	}

	refreshTimer := time.NewTimer(time.Duration(refreshInterval) * time.Second)

	update := func() {

		c, err := awsconfig.GetCredentials(credentialsURI)
		if err != nil {
			fmt.Println("Error getting credentials:", err)
			return
		}

		err = awsconfig.UpdateCredentialsFile(c)
		if err != nil {
			fmt.Println("Error updating credentials file:", err)
			return
		}
	}

	update()

	// kick off timer to refresh credentials
	for {
		select {
		case <-refreshTimer.C:
			fmt.Println("Refreshing credentials")
			update()
			refreshTimer.Reset(time.Duration(refreshInterval) * time.Second)
		}
	}
}
