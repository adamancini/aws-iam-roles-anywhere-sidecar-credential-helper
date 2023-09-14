package main

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	awsconfig "github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper/awsconfig"
)

const defaultListenPort = ":3000"
const defaultCredentialsURI = "http://localhost:8080/creds"
const defaultRefreshInterval = 300 // 5 minutes

func update(uri string) error {
	c, err := awsconfig.GetCredentials(uri)
	if err != nil {
		log.Println("Error getting credentials:", err)
		return err
	}
	err = awsconfig.UpdateCredentialsFile(c)
	if err != nil {
		log.Println("Error updating credentials file:", err)
		return err
	}
	return nil
}

func isValidPort(s string) bool {
	pattern := regexp.MustCompile(`^:\d{1,5}$`)
	return pattern.MatchString(s)
}

func main() {

	listenPort := os.Getenv("LISTEN_PORT")
	if !isValidPort(listenPort) {
		log.Printf("Invalid LISTEN_PORT environment variable. defaulting to %s", defaultListenPort)
		listenPort = ":3000"
	}

	credentialsURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")
	if credentialsURI == "" {
		log.Printf("AWS_CONTAINER_CREDENTIALS_FULL_URI environment variable not set. defaulting to %s", defaultCredentialsURI)
		credentialsURI = defaultCredentialsURI
	}

	refreshInterval := defaultRefreshInterval
	refreshIntervalStr := os.Getenv("AWS_REFRESH_INTERVAL")
	if refreshIntervalStr != "" {
		i, err := strconv.Atoi(refreshIntervalStr)
		if err != nil {
			refreshInterval = i
		} else {
			log.Printf("Invalid AWS_REFRESH_INTERVAL environment variable. defaulting to %d", defaultRefreshInterval)
		}
	}

	err := update(credentialsURI)
	if err != nil {
		log.Println("Error updating credentials:", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(time.Duration(refreshInterval) * time.Second)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	})

	go func() {
		log.Println("Listening on port 3000 for health checks")
		http.ListenAndServe(":3000", nil)
	}()

	go func() {
		for range ticker.C {
			err := update(credentialsURI)
			if err != nil {
				log.Println("Error updating credentials:", err)
			}
			os.Exit(1)
		}
	}()

	select {}
}
