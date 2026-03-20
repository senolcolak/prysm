// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStringSetting(t *testing.T) {
	tests := []struct {
		name         string
		settings     map[string]interface{}
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "key exists",
			settings:     map[string]interface{}{"host": "localhost"},
			key:          "host",
			defaultValue: "default",
			expected:     "localhost",
		},
		{
			name:         "key missing returns default",
			settings:     map[string]interface{}{},
			key:          "host",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "nil settings returns default",
			settings:     nil,
			key:          "host",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "wrong type returns default",
			settings:     map[string]interface{}{"host": 123},
			key:          "host",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty string value",
			settings:     map[string]interface{}{"host": ""},
			key:          "host",
			defaultValue: "default",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringSetting(tt.settings, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntSetting(t *testing.T) {
	tests := []struct {
		name         string
		settings     map[string]interface{}
		key          string
		defaultValue int
		expected     int
	}{
		{
			name:         "key exists",
			settings:     map[string]interface{}{"port": 8080},
			key:          "port",
			defaultValue: 3000,
			expected:     8080,
		},
		{
			name:         "key missing returns default",
			settings:     map[string]interface{}{},
			key:          "port",
			defaultValue: 3000,
			expected:     3000,
		},
		{
			name:         "nil settings returns default",
			settings:     nil,
			key:          "port",
			defaultValue: 3000,
			expected:     3000,
		},
		{
			name:         "wrong type returns default",
			settings:     map[string]interface{}{"port": "8080"},
			key:          "port",
			defaultValue: 3000,
			expected:     3000,
		},
		{
			name:         "zero value",
			settings:     map[string]interface{}{"port": 0},
			key:          "port",
			defaultValue: 3000,
			expected:     0,
		},
		{
			name:         "negative value",
			settings:     map[string]interface{}{"port": -1},
			key:          "port",
			defaultValue: 3000,
			expected:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIntSetting(tt.settings, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBoolSetting(t *testing.T) {
	tests := []struct {
		name         string
		settings     map[string]interface{}
		key          string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "true value",
			settings:     map[string]interface{}{"enabled": true},
			key:          "enabled",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "false value",
			settings:     map[string]interface{}{"enabled": false},
			key:          "enabled",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "key missing returns default true",
			settings:     map[string]interface{}{},
			key:          "enabled",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "key missing returns default false",
			settings:     map[string]interface{}{},
			key:          "enabled",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "nil settings returns default",
			settings:     nil,
			key:          "enabled",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "wrong type returns default",
			settings:     map[string]interface{}{"enabled": "true"},
			key:          "enabled",
			defaultValue: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBoolSetting(tt.settings, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFloat64Setting(t *testing.T) {
	tests := []struct {
		name         string
		settings     map[string]interface{}
		key          string
		defaultValue float64
		expected     float64
	}{
		{
			name:         "float value",
			settings:     map[string]interface{}{"threshold": 80.5},
			key:          "threshold",
			defaultValue: 50.0,
			expected:     80.5,
		},
		{
			name:         "integer as float",
			settings:     map[string]interface{}{"threshold": float64(100)},
			key:          "threshold",
			defaultValue: 50.0,
			expected:     100.0,
		},
		{
			name:         "key missing returns default",
			settings:     map[string]interface{}{},
			key:          "threshold",
			defaultValue: 50.0,
			expected:     50.0,
		},
		{
			name:         "nil settings returns default",
			settings:     nil,
			key:          "threshold",
			defaultValue: 50.0,
			expected:     50.0,
		},
		{
			name:         "wrong type returns default",
			settings:     map[string]interface{}{"threshold": "80.5"},
			key:          "threshold",
			defaultValue: 50.0,
			expected:     50.0,
		},
		{
			name:         "zero value",
			settings:     map[string]interface{}{"threshold": 0.0},
			key:          "threshold",
			defaultValue: 50.0,
			expected:     0.0,
		},
		{
			name:         "negative value",
			settings:     map[string]interface{}{"threshold": -10.5},
			key:          "threshold",
			defaultValue: 50.0,
			expected:     -10.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFloat64Setting(tt.settings, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStringSliceSetting(t *testing.T) {
	tests := []struct {
		name         string
		settings     map[string]interface{}
		key          string
		defaultValue []string
		expected     []string
	}{
		{
			name:         "slice value",
			settings:     map[string]interface{}{"disks": []interface{}{"sda", "sdb", "sdc"}},
			key:          "disks",
			defaultValue: []string{"default"},
			expected:     []string{"sda", "sdb", "sdc"},
		},
		{
			name:         "empty slice",
			settings:     map[string]interface{}{"disks": []interface{}{}},
			key:          "disks",
			defaultValue: []string{"default"},
			expected:     nil, // Empty result from empty interface slice
		},
		{
			name:         "key missing returns default",
			settings:     map[string]interface{}{},
			key:          "disks",
			defaultValue: []string{"sda", "sdb"},
			expected:     []string{"sda", "sdb"},
		},
		{
			name:         "nil settings returns default",
			settings:     nil,
			key:          "disks",
			defaultValue: []string{"sda"},
			expected:     []string{"sda"},
		},
		{
			name:         "wrong type returns default",
			settings:     map[string]interface{}{"disks": "sda,sdb"},
			key:          "disks",
			defaultValue: []string{"default"},
			expected:     []string{"default"},
		},
		{
			name:         "mixed types in slice filters non-strings",
			settings:     map[string]interface{}{"disks": []interface{}{"sda", 123, "sdb"}},
			key:          "disks",
			defaultValue: []string{"default"},
			expected:     []string{"sda", "sdb"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringSliceSetting(tt.settings, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
global:
  nats_url: "nats://localhost:4222"
  admin_url: "http://localhost:8000"
  access_key: "test-access-key"
  secret_key: "test-secret-key"
  node_name: "test-node"
  instance_id: "test-instance"

producers:
  - name: "bucket_notify"
    type: "bucket_notify"
    settings:
      nats_subject: "rgw.notifications"
      endpoint_port: 9090
  - name: "resource_usage"
    type: "resource_usage"
    settings:
      prometheus: true
      interval: 60
      disks:
        - sda
        - sdb
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify global config
	assert.Equal(t, "nats://localhost:4222", config.Global.NatsURL)
	assert.Equal(t, "http://localhost:8000", config.Global.AdminURL)
	assert.Equal(t, "test-access-key", config.Global.AccessKey)
	assert.Equal(t, "test-secret-key", config.Global.SecretKey)
	assert.Equal(t, "test-node", config.Global.NodeName)
	assert.Equal(t, "test-instance", config.Global.InstanceID)

	// Verify producers
	assert.Len(t, config.Producers, 2)

	// First producer
	assert.Equal(t, "bucket_notify", config.Producers[0].Name)
	assert.Equal(t, "bucket_notify", config.Producers[0].Type)

	// Second producer
	assert.Equal(t, "resource_usage", config.Producers[1].Name)
	assert.Equal(t, "resource_usage", config.Producers[1].Type)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	config, err := LoadConfig("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestGlobalConfig_StructFields(t *testing.T) {
	config := GlobalConfig{
		NatsURL:    "nats://localhost:4222",
		AdminURL:   "http://localhost:8000",
		AccessKey:  "access",
		SecretKey:  "secret",
		NodeName:   "node1",
		InstanceID: "instance1",
	}

	assert.Equal(t, "nats://localhost:4222", config.NatsURL)
	assert.Equal(t, "http://localhost:8000", config.AdminURL)
	assert.Equal(t, "access", config.AccessKey)
	assert.Equal(t, "secret", config.SecretKey)
	assert.Equal(t, "node1", config.NodeName)
	assert.Equal(t, "instance1", config.InstanceID)
}

func TestProducerConfig_StructFields(t *testing.T) {
	config := ProducerConfig{
		Name: "test-producer",
		Type: "bucket_notify",
		Settings: map[string]interface{}{
			"port":    8080,
			"enabled": true,
		},
	}

	assert.Equal(t, "test-producer", config.Name)
	assert.Equal(t, "bucket_notify", config.Type)
	assert.Equal(t, 8080, config.Settings["port"])
	assert.Equal(t, true, config.Settings["enabled"])
}
