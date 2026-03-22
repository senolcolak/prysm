// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package opslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "408 Request Timeout",
			status:   "408",
			expected: true,
		},
		{
			name:     "504 Gateway Timeout",
			status:   "504",
			expected: true,
		},
		{
			name:     "598 Network Read Timeout",
			status:   "598",
			expected: true,
		},
		{
			name:     "499 Client Closed Request",
			status:   "499",
			expected: true,
		},
		{
			name:     "200 OK - not timeout",
			status:   "200",
			expected: false,
		},
		{
			name:     "400 Bad Request - not timeout",
			status:   "400",
			expected: false,
		},
		{
			name:     "404 Not Found - not timeout",
			status:   "404",
			expected: false,
		},
		{
			name:     "500 Internal Server Error - not timeout",
			status:   "500",
			expected: false,
		},
		{
			name:     "502 Bad Gateway - not timeout",
			status:   "502",
			expected: false,
		},
		{
			name:     "503 Service Unavailable - not timeout",
			status:   "503",
			expected: false,
		},
		{
			name:     "empty string",
			status:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTimeoutError(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTimeoutType(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "408 Request Timeout",
			status:   "408",
			expected: "request_timeout",
		},
		{
			name:     "504 Gateway Timeout",
			status:   "504",
			expected: "gateway_timeout",
		},
		{
			name:     "598 Network Read Timeout",
			status:   "598",
			expected: "network_read_timeout",
		},
		{
			name:     "499 Client Closed Request",
			status:   "499",
			expected: "client_closed_request",
		},
		{
			name:     "unknown status",
			status:   "500",
			expected: "unknown_timeout",
		},
		{
			name:     "empty status",
			status:   "",
			expected: "unknown_timeout",
		},
		{
			name:     "200 OK",
			status:   "200",
			expected: "unknown_timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTimeoutType(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCategorizeHTTPError(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		// Timeout errors
		{
			name:     "408 Request Timeout - timeout category",
			status:   "408",
			expected: "timeout",
		},
		{
			name:     "504 Gateway Timeout - timeout category",
			status:   "504",
			expected: "timeout",
		},
		{
			name:     "598 Network Read Timeout - timeout category",
			status:   "598",
			expected: "timeout",
		},
		{
			name:     "499 Client Closed Request - timeout category",
			status:   "499",
			expected: "timeout",
		},
		// Connection errors
		{
			name:     "502 Bad Gateway - connection category",
			status:   "502",
			expected: "connection",
		},
		{
			name:     "503 Service Unavailable - connection category",
			status:   "503",
			expected: "connection",
		},
		// Client errors (4xx)
		{
			name:     "400 Bad Request - client category",
			status:   "400",
			expected: "client",
		},
		{
			name:     "401 Unauthorized - client category",
			status:   "401",
			expected: "client",
		},
		{
			name:     "403 Forbidden - client category",
			status:   "403",
			expected: "client",
		},
		{
			name:     "404 Not Found - client category",
			status:   "404",
			expected: "client",
		},
		{
			name:     "405 Method Not Allowed - client category",
			status:   "405",
			expected: "client",
		},
		{
			name:     "409 Conflict - client category",
			status:   "409",
			expected: "client",
		},
		{
			name:     "429 Too Many Requests - client category",
			status:   "429",
			expected: "client",
		},
		// Server errors (5xx) - not timeout or connection
		{
			name:     "500 Internal Server Error - server category",
			status:   "500",
			expected: "server",
		},
		{
			name:     "501 Not Implemented - server category",
			status:   "501",
			expected: "server",
		},
		{
			name:     "505 HTTP Version Not Supported - server category",
			status:   "505",
			expected: "server",
		},
		// Unknown
		{
			name:     "empty status - unknown category",
			status:   "",
			expected: "unknown",
		},
		{
			name:     "200 OK - unknown category",
			status:   "200",
			expected: "unknown",
		},
		{
			name:     "201 Created - unknown category",
			status:   "201",
			expected: "unknown",
		},
		{
			name:     "204 No Content - unknown category",
			status:   "204",
			expected: "unknown",
		},
		{
			name:     "301 Moved Permanently - unknown category",
			status:   "301",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeHTTPError(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorCategorization_Consistency(t *testing.T) {
	// Test that all timeout errors are correctly categorized
	timeoutStatuses := []string{"408", "504", "598", "499"}
	for _, status := range timeoutStatuses {
		t.Run("timeout_"+status, func(t *testing.T) {
			assert.True(t, IsTimeoutError(status), "Should be timeout error")
			assert.Equal(t, "timeout", CategorizeHTTPError(status), "Should be categorized as timeout")
			assert.NotEqual(t, "unknown_timeout", GetTimeoutType(status), "Should have specific timeout type")
		})
	}

	// Test that non-timeout 4xx errors are client errors
	clientStatuses := []string{"400", "401", "403", "404", "405", "409", "429"}
	for _, status := range clientStatuses {
		t.Run("client_"+status, func(t *testing.T) {
			assert.False(t, IsTimeoutError(status), "Should not be timeout error")
			assert.Equal(t, "client", CategorizeHTTPError(status), "Should be categorized as client")
		})
	}

	// Test that 502/503 are connection errors, not server errors
	connectionStatuses := []string{"502", "503"}
	for _, status := range connectionStatuses {
		t.Run("connection_"+status, func(t *testing.T) {
			assert.False(t, IsTimeoutError(status), "Should not be timeout error")
			assert.Equal(t, "connection", CategorizeHTTPError(status), "Should be categorized as connection")
		})
	}
}
