package awsconfig

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestToCredentialsINI(t *testing.T) {
	c := &awsConfig{
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Token:           "session-token-value",
	}

	got := c.toCredentialsINI()
	want := "[default]\naws_access_key_id = AKIAIOSFODNN7EXAMPLE\naws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\naws_session_token = session-token-value\n"

	if got != want {
		t.Errorf("toCredentialsINI() =\n%s\nwant:\n%s", got, want)
	}
}

func TestToConfigINI(t *testing.T) {
	c := &awsConfig{}
	got := c.toConfigINI()
	want := "[default]\n"

	if got != want {
		t.Errorf("toConfigINI() = %q, want %q", got, want)
	}
}

func TestGetCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"AccessKeyId": "AKIAIOSFODNN7EXAMPLE",
			"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"Token": "session-token",
			"Expiration": "2024-01-01T00:00:00Z",
			"RoleArn": "arn:aws:iam::123456789012:role/test-role"
		}`))
	}))
	defer server.Close()

	c, err := GetCredentials(server.URL)
	if err != nil {
		t.Fatalf("GetCredentials() error = %v", err)
	}

	if c.AccessKeyId != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("AccessKeyId = %q, want %q", c.AccessKeyId, "AKIAIOSFODNN7EXAMPLE")
	}
	if c.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("SecretAccessKey = %q, want %q", c.SecretAccessKey, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	}
	if c.Token != "session-token" {
		t.Errorf("Token = %q, want %q", c.Token, "session-token")
	}
	if c.Expiration != "2024-01-01T00:00:00Z" {
		t.Errorf("Expiration = %q, want %q", c.Expiration, "2024-01-01T00:00:00Z")
	}
	if c.RoleArn != "arn:aws:iam::123456789012:role/test-role" {
		t.Errorf("RoleArn = %q, want %q", c.RoleArn, "arn:aws:iam::123456789012:role/test-role")
	}
}

func TestGetCredentials_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	_, err := GetCredentials(server.URL)
	if err == nil {
		t.Error("GetCredentials() expected error for invalid JSON, got nil")
	}
}

func TestGetCredentials_ConnectionError(t *testing.T) {
	_, err := GetCredentials("http://localhost:1")
	if err == nil {
		t.Error("GetCredentials() expected error for connection failure, got nil")
	}
}

func TestUpdateCredentialsFile(t *testing.T) {
	tmpDir := t.TempDir()
	awsDir := filepath.Join(tmpDir, ".aws")
	if err := os.MkdirAll(awsDir, 0755); err != nil {
		t.Fatalf("failed to create .aws dir: %v", err)
	}

	orig := getHomeDir
	getHomeDir = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { getHomeDir = orig })

	c := &awsConfig{
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Token:           "session-token",
	}

	if err := UpdateCredentialsFile(c); err != nil {
		t.Fatalf("UpdateCredentialsFile() error = %v", err)
	}

	credBytes, err := os.ReadFile(filepath.Join(awsDir, "credentials"))
	if err != nil {
		t.Fatalf("failed to read credentials file: %v", err)
	}
	wantCreds := "[default]\naws_access_key_id = AKIAIOSFODNN7EXAMPLE\naws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\naws_session_token = session-token\n"
	if string(credBytes) != wantCreds {
		t.Errorf("credentials file =\n%s\nwant:\n%s", string(credBytes), wantCreds)
	}

	configBytes, err := os.ReadFile(filepath.Join(awsDir, "config"))
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	wantConfig := "[default]\n"
	if string(configBytes) != wantConfig {
		t.Errorf("config file = %q, want %q", string(configBytes), wantConfig)
	}
}

func TestUpdateCredentialsFile_NoTempFilesLeftBehind(t *testing.T) {
	tmpDir := t.TempDir()
	awsDir := filepath.Join(tmpDir, ".aws")
	if err := os.MkdirAll(awsDir, 0755); err != nil {
		t.Fatalf("failed to create .aws dir: %v", err)
	}

	orig := getHomeDir
	getHomeDir = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { getHomeDir = orig })

	c := &awsConfig{
		AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "secret",
		Token:           "token",
	}

	if err := UpdateCredentialsFile(c); err != nil {
		t.Fatalf("UpdateCredentialsFile() error = %v", err)
	}

	entries, err := os.ReadDir(awsDir)
	if err != nil {
		t.Fatalf("failed to read .aws dir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestUpdateCredentialsFile_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	awsDir := filepath.Join(tmpDir, ".aws")
	if err := os.MkdirAll(awsDir, 0755); err != nil {
		t.Fatalf("failed to create .aws dir: %v", err)
	}

	orig := getHomeDir
	getHomeDir = func() (string, error) { return tmpDir, nil }
	t.Cleanup(func() { getHomeDir = orig })

	// Write initial credentials
	first := &awsConfig{
		AccessKeyId:     "OLD_KEY",
		SecretAccessKey: "OLD_SECRET",
		Token:           "OLD_TOKEN",
	}
	if err := UpdateCredentialsFile(first); err != nil {
		t.Fatalf("first UpdateCredentialsFile() error = %v", err)
	}

	// Overwrite with new credentials
	second := &awsConfig{
		AccessKeyId:     "NEW_KEY",
		SecretAccessKey: "NEW_SECRET",
		Token:           "NEW_TOKEN",
	}
	if err := UpdateCredentialsFile(second); err != nil {
		t.Fatalf("second UpdateCredentialsFile() error = %v", err)
	}

	credBytes, err := os.ReadFile(filepath.Join(awsDir, "credentials"))
	if err != nil {
		t.Fatalf("failed to read credentials file: %v", err)
	}

	wantCreds := "[default]\naws_access_key_id = NEW_KEY\naws_secret_access_key = NEW_SECRET\naws_session_token = NEW_TOKEN\n"
	if string(credBytes) != wantCreds {
		t.Errorf("credentials file after overwrite =\n%s\nwant:\n%s", string(credBytes), wantCreds)
	}
}
