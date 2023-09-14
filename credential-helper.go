package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	awsconfig "github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper/awsconfig"
)

func main() {

	credentialsURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")
	if credentialsURI == "" {
		log.Println("AWS_CONTAINER_CREDENTIALS_FULL_URI environment variable not set. defaulting to localhost:8080/creds")
		credentialsURI = "http://localhost:8080/creds"
	}

	refreshIntervalStr := os.Getenv("AWS_REFRESH_INTERVAL")
	refreshInterval := 300 // Default to 5 minutes (300 seconds)
	if refreshIntervalStr != "" {
		interval, err := strconv.Atoi(refreshIntervalStr)
		if err != nil {
			log.Println("Invalid AWS_REFRESH_INTERVAL:", err)
			return
		}
		refreshInterval = interval
	}

	ticker := time.NewTicker(time.Duration(refreshInterval) * time.Second)

	update := func() {

		c, err := awsconfig.GetCredentials(credentialsURI)
		if err != nil {
			log.Println("Error getting credentials:", err)
			return
		}

		err = awsconfig.UpdateCredentialsFile(c)
		if err != nil {
			log.Println("Error updating credentials file:", err)
			return
		}
	}

	update()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	})

	go func() {
		for range ticker.C {
			update()
		}
	}()

	go func() {
		log.Println("Listening on port 8080 for health checks")
		http.ListenAndServe(":8080", nil)
	}()

	select {}
}
