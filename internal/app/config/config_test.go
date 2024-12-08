package config

import (
	"testing"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		want *Config
		name string
	}{
		{name: "positive test #1", want: &Config{URI: ":8080"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := Setup()
			if c.URI != tt.want.URI {
				t.Errorf("Setup() returns bad config")
			}
		})
	}
}
