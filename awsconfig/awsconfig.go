package awsconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

const awsConfigFileName = "config"
const awsCredentialsFileName = "credentials"
const awsConfigDir = ".aws"

type awsConfig struct {
	AccessKeyId     string `json:"AccessKeyId"`
	Expiration      string `json:"Expiration"`
	RoleArn         string `json:"RoleArn"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Token           string `json:"Token"`
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err == nil {
		return usr.HomeDir, nil
	}
	home := os.Getenv("HOME")
	if home != "" {
		return home, nil
	}
	return "", fmt.Errorf("user home directory not found")
}

func GetCredentials(uri string) (*awsConfig, error) {
	resp, err := http.Get(uri)
	if err != nil {
		log.Println("Error making HTTP request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}

	var awsConfig awsConfig
	err = json.Unmarshal(body, &awsConfig)
	if err != nil {
		log.Println("Error parsing JSON:", err)
		return nil, err
	}

	return &awsConfig, nil
}

func (c *awsConfig) toCredentialsINI() string {
	return fmt.Sprintf("[default]\naws_access_key_id = %s\naws_secret_access_key = %s\naws_session_token = %s\n", c.AccessKeyId, c.SecretAccessKey, c.Token)
}

func (c *awsConfig) toConfigINI() string {
	return fmt.Sprintf("[default]\nrole_arn = %s\nsource_profile = default\n", c.RoleArn)
}

// atomically update ~/.aws/credentials and ~/.aws/config
func UpdateCredentialsFile(c *awsConfig) error {
	awsConfig := c.toConfigINI()
	awsCredentials := c.toCredentialsINI()

	homeDir, err := getHomeDir()
	if err != nil {
		return fmt.Errorf("Error getting home directory: %w", err)
	}

	configFilePath := filepath.Join(homeDir, awsConfigDir, awsConfigFileName)
	credentialsFilePath := filepath.Join(homeDir, awsConfigDir, awsCredentialsFileName)

	tempConfigFilePath := configFilePath + ".tmp"
	tempCredentialsFilePath := credentialsFilePath + ".tmp"

	err = os.WriteFile(tempConfigFilePath, []byte(awsConfig), 0644)
	if err != nil {
		return fmt.Errorf("Error writing config file: %w", err)
	}

	err = os.WriteFile(tempCredentialsFilePath, []byte(awsCredentials), 0644)
	if err != nil {
		return fmt.Errorf("Error writing credentials file: %w", err)
	}

	err = os.Rename(tempConfigFilePath, configFilePath)
	if err != nil {
		return fmt.Errorf("Error renaming config file: %w", err)
	}

	err = os.Rename(tempCredentialsFilePath, credentialsFilePath)
	if err != nil {
		return fmt.Errorf("Error renaming credentials file: %w", err)
	}

	log.Println("Updated aws config successfully")
	return nil
}
