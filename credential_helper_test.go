package main

import "testing"

func TestParseRefreshInterval(t *testing.T) {
	tests := []struct {
		name string
		input string
		want int
	}{
		{"valid integer", "900", 900},
		{"valid small value", "60", 60},
		{"empty string uses default", "", defaultRefreshInterval},
		{"non-numeric uses default", "abc", defaultRefreshInterval},
		{"negative value", "-1", -1},
		{"zero", "0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRefreshInterval(tt.input)
			if got != tt.want {
				t.Errorf("parseRefreshInterval(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid port 3000", ":3000", true},
		{"valid port 80", ":80", true},
		{"valid port 443", ":443", true},
		{"valid port 65535", ":65535", true},
		{"valid single digit", ":1", true},
		{"missing colon prefix", "3000", false},
		{"empty string", "", false},
		{"letters after colon", ":abc", false},
		{"just colon", ":", false},
		{"colon with space", ": 3000", false},
		{"too many digits", ":123456", false},
		{"negative", ":-1", false},
		{"zero", ":0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPort(tt.input)
			if got != tt.want {
				t.Errorf("isValidPort(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
