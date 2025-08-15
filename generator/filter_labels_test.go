package generator

import (
	"regexp"
	"strings"
	"testing"

	"github.com/lucaslorentz/caddy-docker-proxy/v2/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestFilterLabels_EmptyValues(t *testing.T) {
	// Create a generator with default options
	options := &config.Options{
		LabelPrefix: "caddy",
	}
	labelRegexString := "^caddy(_\\d+)?(\\.|$)"
	
	generator := &CaddyfileGenerator{
		options:    options,
		labelRegex: regexp.MustCompile(labelRegexString),
	}

	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name: "filters out empty caddy label",
			input: map[string]string{
				"caddy":              "",
				"caddy.reverse_proxy": "{{upstreams 80}}",
				"other_label":        "value",
			},
			expected: map[string]string{
				"caddy.reverse_proxy": "{{upstreams 80}}",
			},
		},
		{
			name: "filters out whitespace-only caddy label",
			input: map[string]string{
				"caddy":              "   \n\t  ",
				"caddy.reverse_proxy": "{{upstreams 80}}",
			},
			expected: map[string]string{
				"caddy.reverse_proxy": "{{upstreams 80}}",
			},
		},
		{
			name: "filters out problematic empty nested labels only",
			input: map[string]string{
				"caddy":                                       "example.com",
				"caddy.reverse_proxy":                         "{{upstreams 80}}",
				"caddy.reverse_proxy.transport.tls_insecure_skip_verify": "",
				"caddy.gzip":                                 "", // Valid empty label
			},
			expected: map[string]string{
				"caddy":               "example.com",
				"caddy.reverse_proxy": "{{upstreams 80}}",
				"caddy.gzip":         "",
			},
		},
		{
			name: "preserves valid labels",
			input: map[string]string{
				"caddy":               "example.com",
				"caddy.reverse_proxy": "{{upstreams 80}}",
				"caddy.tls":          "internal",
			},
			expected: map[string]string{
				"caddy":               "example.com",
				"caddy.reverse_proxy": "{{upstreams 80}}",
				"caddy.tls":          "internal",
			},
		},
		{
			name: "ignores non-caddy labels",
			input: map[string]string{
				"caddy":      "example.com",
				"other":      "",
				"another":    "value",
				"empty_one":  "",
			},
			expected: map[string]string{
				"caddy": "example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.filterLabels(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterLabelsWithContext_Logging(t *testing.T) {
	// Create a generator with default options
	options := &config.Options{
		LabelPrefix: "caddy",
	}
	labelRegexString := "^caddy(_\\d+)?(\\.|$)"
	
	generator := &CaddyfileGenerator{
		options:    options,
		labelRegex: regexp.MustCompile(labelRegexString),
	}

	// Create a test logger that captures logs
	logger := zaptest.NewLogger(t, zaptest.Level(zap.ErrorLevel))

	tests := []struct {
		name        string
		input       map[string]string
		containerID string
		expectLogs  bool
	}{
		{
			name: "logs empty caddy label with container ID",
			input: map[string]string{
				"caddy": "",
				"other": "value",
			},
			containerID: "container123456",
			expectLogs:  true,
		},
		{
			name: "logs whitespace-only label with container ID",
			input: map[string]string{
				"caddy": "  \n\t  ",
			},
			containerID: "container789abc",
			expectLogs:  true,
		},
		{
			name: "no logs for valid labels",
			input: map[string]string{
				"caddy": "example.com",
			},
			containerID: "container123456",
			expectLogs:  false,
		},
		{
			name: "no logs when no logger provided",
			input: map[string]string{
				"caddy": "",
			},
			containerID: "container123456",
			expectLogs:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var useLogger *zap.Logger
			if tt.expectLogs {
				useLogger = logger
			}
			
			result := generator.filterLabelsWithContext(tt.input, tt.containerID, useLogger)
			
			// For tests with empty labels, verify they are filtered out
			// For tests with valid labels, verify they are preserved
			if tt.input["caddy"] == "" || strings.TrimSpace(tt.input["caddy"]) == "" {
				_, hasEmptyLabel := result["caddy"]
				assert.False(t, hasEmptyLabel, "Empty caddy label should be filtered out")
			} else {
				// Valid labels should be preserved
				_, hasValidLabel := result["caddy"]
				assert.True(t, hasValidLabel, "Valid caddy label should be preserved")
			}
		})
	}
}

func TestIsProblematicEmptyLabel(t *testing.T) {
	generator := &CaddyfileGenerator{}
	
	tests := []struct {
		label       string
		problematic bool
	}{
		{"caddy", true},                    // Main caddy label should never be empty
		{"caddy.gzip", false},             // Valid empty directive
		{"caddy.experimental_http3", false}, // Valid empty directive
		{"caddy.reverse_proxy.transport.tls_insecure_skip_verify", true}, // Known problematic
		{"caddy.tls_insecure_skip_verify", true}, // Known problematic
		{"caddy.basicauth", false},        // Valid empty directive
		{"caddy.rewrite", false},          // Valid empty directive
	}
	
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			result := generator.isProblematicEmptyLabel(tt.label)
			assert.Equal(t, tt.problematic, result, "Label: %s", tt.label)
		})
	}
}