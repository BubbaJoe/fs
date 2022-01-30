package client_test

import (
	"fs-store/client"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ClientInit_DomainCheck(t *testing.T) {
	for _, testCase := range []struct {
		actual      string
		expected    string
		description string
	}{
		{"http://domain.com", "http://domain.com", "HTTP with domain"},
		{"https://domain", "https://domain", "HTTPS with domain"},
		{"http://domain/", "http://domain", "HTTP with domain with trailing slash"},
		{"http://domain/path", "http://domain", "HTTP with domain with path"},
		{"http://domain/path?q=v", "http://domain", "HTTP with domain with path and query string"},
		{"http://domain:8080", "http://domain:8080", "HTTP with domain with port"},
		{"http://localhost:8080", "http://localhost:8080", "HTTP with localhost and port"},
		{"http://192.168.0.1", "http://192.168.0.1", "HTTP with IP"},
		{"http://192.168.0.1:8080", "http://192.168.0.1:8080", "HTTP with IP and port"},
		{"http://::1", "http://::1", "HTTP with IPv6"},
		{"http://::1:8080", "http://::1:8080", "HTTP with IPv6"},
		// No scheme
		{"::1", "http://::1", "IPv6"},
		{"::1:8080", "http://::1:8080", "IPv6 with port"},
		{"127.0.0.1", "http://127.0.0.1", "IP"},
		{"127.0.0.1:8080", "http://127.0.0.1:8080", "IP with port"},
		{"domain.com", "http://domain.com", "domain"},
		{"domain.com:8080", "http://domain.com:8080", "domain with port"},
		// No scheme
		{"::1", "https://::1", "IPv6"},
		{"::1:8080", "https://::1:8080", "IPv6 with port"},
		{"127.0.0.1", "https://127.0.0.1", "IP"},
		{"127.0.0.1:8080", "https://127.0.0.1:8080", "IP with port"},
		{"domain.com", "https://domain.com", "domain"},
		{"domain.com:8080", "https://domain.com:8080", "domain with port"},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			conf, err := client.NewFSClientConfig(testCase.actual,
				strings.HasPrefix(testCase.expected, "https"))

			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, testCase.expected, conf.Client.BaseURL,
				"Expected domain to be %s, got %s", testCase.expected, conf.Client.BaseURL)
		})
	}
}
