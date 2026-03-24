// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	key := "TEST_KEY"
	fallback := "default_value"

	// Test when the environment variable is not set
	value := getEnv(key, fallback)
	assert.Equal(t, fallback, value)

	// Test when the environment variable is set
	expectedValue := "expected_value"
	os.Setenv(key, expectedValue)
	value = getEnv(key, fallback)
	assert.Equal(t, expectedValue, value)

	// Clean up
	os.Unsetenv(key)
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
		setEnv       bool
	}{
		{
			name:         "valid int",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
			setEnv:       true,
		},
		{
			name:         "invalid int",
			envValue:     "not_a_number",
			defaultValue: 10,
			expected:     10,
			setEnv:       true,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
			setEnv:       true,
		},
		{
			name:         "env not set",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
			setEnv:       false,
		},
		{
			name:         "negative int",
			envValue:     "-5",
			defaultValue: 10,
			expected:     -5,
			setEnv:       true,
		},
		{
			name:         "zero",
			envValue:     "0",
			defaultValue: 10,
			expected:     0,
			setEnv:       true,
		},
	}

	key := "TEST_INT_KEY"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(key)
			if tt.setEnv {
				os.Setenv(key, tt.envValue)
			}
			result := getEnvInt(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
			os.Unsetenv(key)
		})
	}
}

func TestGetEnvInt64(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int64
		expected     int64
		setEnv       bool
	}{
		{
			name:         "valid int64",
			envValue:     "9223372036854775807", // Max int64
			defaultValue: 10,
			expected:     9223372036854775807,
			setEnv:       true,
		},
		{
			name:         "invalid int64",
			envValue:     "not_a_number",
			defaultValue: 10,
			expected:     10,
			setEnv:       true,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
			setEnv:       true,
		},
		{
			name:         "env not set",
			envValue:     "",
			defaultValue: 10,
			expected:     10,
			setEnv:       false,
		},
		{
			name:         "negative int64",
			envValue:     "-9223372036854775808", // Min int64
			defaultValue: 10,
			expected:     -9223372036854775808,
			setEnv:       true,
		},
	}

	key := "TEST_INT64_KEY"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(key)
			if tt.setEnv {
				os.Setenv(key, tt.envValue)
			}
			result := getEnvInt64(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
			os.Unsetenv(key)
		})
	}
}

func TestGetEnvFloat(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue float64
		expected     float64
		setEnv       bool
	}{
		{
			name:         "valid float",
			envValue:     "3.14159",
			defaultValue: 0.0,
			expected:     3.14159,
			setEnv:       true,
		},
		{
			name:         "invalid float",
			envValue:     "not_a_number",
			defaultValue: 1.5,
			expected:     1.5,
			setEnv:       true,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: 1.5,
			expected:     1.5,
			setEnv:       true,
		},
		{
			name:         "env not set",
			envValue:     "",
			defaultValue: 1.5,
			expected:     1.5,
			setEnv:       false,
		},
		{
			name:         "negative float",
			envValue:     "-2.5",
			defaultValue: 1.5,
			expected:     -2.5,
			setEnv:       true,
		},
		{
			name:         "integer as float",
			envValue:     "42",
			defaultValue: 1.5,
			expected:     42.0,
			setEnv:       true,
		},
		{
			name:         "scientific notation",
			envValue:     "1.5e10",
			defaultValue: 0.0,
			expected:     1.5e10,
			setEnv:       true,
		},
	}

	key := "TEST_FLOAT_KEY"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(key)
			if tt.setEnv {
				os.Setenv(key, tt.envValue)
			}
			result := getEnvFloat(key, tt.defaultValue)
			assert.InDelta(t, tt.expected, result, 0.00001)
			os.Unsetenv(key)
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
		setEnv       bool
	}{
		{
			name:         "true value",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "false value",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "1 as true",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "0 as false",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "invalid value",
			envValue:     "not_a_bool",
			defaultValue: true,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: true,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "env not set",
			envValue:     "",
			defaultValue: false,
			expected:     false,
			setEnv:       false,
		},
		{
			name:         "TRUE uppercase",
			envValue:     "TRUE",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "False mixed case",
			envValue:     "False",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
	}

	key := "TEST_BOOL_KEY"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(key)
			if tt.setEnv {
				os.Setenv(key, tt.envValue)
			}
			result := getEnvBool(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
			os.Unsetenv(key)
		})
	}
}

func TestGetEnvInt64Slice(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue []int64
		expected     []int64
		setEnv       bool
	}{
		{
			name:         "valid slice",
			envValue:     "1,2,3,4,5",
			defaultValue: []int64{0},
			expected:     []int64{1, 2, 3, 4, 5},
			setEnv:       true,
		},
		{
			name:         "single value",
			envValue:     "42",
			defaultValue: []int64{0},
			expected:     []int64{42},
			setEnv:       true,
		},
		{
			name:         "invalid value in slice",
			envValue:     "1,2,invalid,4",
			defaultValue: []int64{10, 20},
			expected:     []int64{10, 20},
			setEnv:       true,
		},
		{
			name:         "empty value",
			envValue:     "",
			defaultValue: []int64{10, 20},
			expected:     []int64{10, 20},
			setEnv:       true,
		},
		{
			name:         "env not set",
			envValue:     "",
			defaultValue: []int64{10, 20},
			expected:     []int64{10, 20},
			setEnv:       false,
		},
		{
			name:         "negative values",
			envValue:     "-1,-2,-3",
			defaultValue: []int64{0},
			expected:     []int64{-1, -2, -3},
			setEnv:       true,
		},
		{
			name:         "large int64 values",
			envValue:     "9223372036854775807,-9223372036854775808",
			defaultValue: []int64{0},
			expected:     []int64{9223372036854775807, -9223372036854775808},
			setEnv:       true,
		},
	}

	key := "TEST_INT64_SLICE_KEY"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(key)
			if tt.setEnv {
				os.Setenv(key, tt.envValue)
			}
			result := getEnvInt64Slice(key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
			os.Unsetenv(key)
		})
	}
}

func TestSetUpLogs(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		shouldErr bool
	}{
		{
			name:      "debug level",
			level:     "debug",
			shouldErr: false,
		},
		{
			name:      "info level",
			level:     "info",
			shouldErr: false,
		},
		{
			name:      "warn level",
			level:     "warn",
			shouldErr: false,
		},
		{
			name:      "error level",
			level:     "error",
			shouldErr: false,
		},
		{
			name:      "invalid level",
			level:     "invalid_level",
			shouldErr: true,
		},
		{
			name:      "trace level",
			level:     "trace",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setUpLogs(tt.level)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckIfRunningInPod_NotInPod(t *testing.T) {
	// When not in a Kubernetes pod, the function should return false
	// Clean any env vars that might be set
	os.Unsetenv("KUBERNETES_SERVICE_HOST")
	os.Unsetenv("KUBERNETES_SERVICE_PORT")

	// The function checks for files and env vars, so in test env it should return false
	result := checkIfRunningInPod()
	assert.False(t, result)
}
