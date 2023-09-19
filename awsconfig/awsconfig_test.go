package awsconfig

import (
	"testing"
)

func TestAWSConfigToConfigINI(t *testing.T) {

	c := &awsConfig{} 

	result := c.toConfigINI()
	expected := "[default]\n"

	if result != expected {
		t.Errorf("c.toConfigINI() returned %s, expected %s", result, expected)
	}
}
