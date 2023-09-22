package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	awsconfig "github.com/adamancini/aws-iam-roles-anywhere-sidecar-credential-helper/awsconfig"
	helper "github.com/aws/rolesanywhere-credential-helper/aws_signing_helper"
)

const defaultListenPort = ":8080"
const defaultRefreshInterval = 300 // 5 minutes

type CredentialResponse struct {
	AccessKeyID     string `json:"AccessKeyId"`
	Expiration      string
	RoleArn         string
	SecretAccessKey string
	Token           string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func getBoolEnv(key string) bool {
	val, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return false
	}
	return val
}

func getIntEnv(key string, defalt int) int {
	val, err := strconv.ParseInt(os.Getenv(key), 10, 0)
	if err != nil {
		return defalt
	}
	return int(val)
}

func isValidPort(s string) bool {
	pattern := regexp.MustCompile(`^:\d{1,5}$`)
	return pattern.MatchString(s)
}

func update(string) error {
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

func main() {

	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds)

	credentialsOptions := helper.CredentialsOpts{
		PrivateKeyId:        os.Getenv("PRIVATE_KEY_ID"),
		CertificateId:       os.Getenv("CERTIFICATE_ID"),
		CertificateBundleId: os.Getenv("CERTIFICATE_BUNDLE_ID"),
		RoleArn:             os.Getenv("ROLE_ARN"),
		ProfileArnStr:       os.Getenv("PROFILE_ARN"),
		TrustAnchorArnStr:   os.Getenv("TRUST_ANCHOR_ID"),
		SessionDuration:     getIntEnv("SESSION_DURATION", 3600),
		Region:              os.Getenv("AWS_REGION"),
		Endpoint:            os.Getenv("ENDPOINT"),
		NoVerifySSL:         getBoolEnv("NO_VERIFY_SSL"),
		WithProxy:           getBoolEnv("WITH_PROXY"),
		Debug:               getBoolEnv("DEBUG"),
		Version:             os.Getenv("CREDENTIAL_VERSION"),
	}

	listen, err := net.ResolveTCPAddr("tcp", os.Getenv("LISTEN"))
	if err != nil {
		log.Printf("failed to resolve listen address: %v; defaulting to 0.0.0.0:8080", err)
		listen = &net.TCPAddr{IP: net.IPv4zero, Port: 8080}
	}

	// credentialsURI := os.Getenv("AWS_CONTAINER_CREDENTIALS_FULL_URI")
	// if credentialsURI == "" {
	// 	log.Printf("AWS_CONTAINER_CREDENTIALS_FULL_URI environment variable not set. defaulting to %s", defaultCredentialsURI)
	// 	credentialsURI = defaultCredentialsURI
	// }

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

	err = update(listen.String())
	if err != nil {
		log.Println("Error updating credentials:", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(time.Duration(refreshInterval) * time.Second)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	})

	http.HandleFunc("/creds", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		output, err := helper.GenerateCredentials(&credentialsOptions)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&ErrorResponse{Error: err.Error()})
			return
		}
		json.NewEncoder(w).Encode(&CredentialResponse{
			AccessKeyID:     output.AccessKeyId,
			Expiration:      output.Expiration,
			RoleArn:         credentialsOptions.RoleArn,
			SecretAccessKey: output.SecretAccessKey,
			Token:           output.SessionToken,
		})
	})

	go func() {
		for range ticker.C {
			err := update(listen.String())
			if err != nil {
				log.Println("Error updating credentials:", err)
			}
			os.Exit(1)
		}
	}()

	go func() {
		log.Printf("listening on %v\n", listen)
		log.Fatal(http.ListenAndServe(listen.String(), requestLogger(http.DefaultServeMux)))
	}()
}
