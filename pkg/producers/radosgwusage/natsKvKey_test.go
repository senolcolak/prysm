// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package radosgwusage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeUserTenant(t *testing.T) {
	tests := []struct {
		name       string
		user       string
		tenant     string
		wantUser   string
		wantTenant string
	}{
		{
			name:       "plain user no tenant",
			user:       "alice",
			tenant:     "",
			wantUser:   "alice",
			wantTenant: "",
		},
		{
			name:       "user embeds tenant",
			user:       "alice$t1",
			tenant:     "",
			wantUser:   "alice",
			wantTenant: "t1",
		},
		{
			name:       "explicit tenant wins",
			user:       "alice$t1",
			tenant:     "t2",
			wantUser:   "alice",
			wantTenant: "t2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, gotTenant := NormalizeUserTenant(tt.user, tt.tenant)
			if gotUser != tt.wantUser || gotTenant != tt.wantTenant {
				t.Fatalf("NormalizeUserTenant(%q, %q) = (%q, %q), want (%q, %q)",
					tt.user, tt.tenant, gotUser, gotTenant, tt.wantUser, tt.wantTenant)
			}
		})
	}
}

func TestBucketAndUsageKeyParity(t *testing.T) {
	tests := []struct {
		name       string
		owner      string
		tenant     string
		usageUser  string
		bucketName string
	}{
		{
			name:       "owner has inline tenant",
			owner:      "alice$t1",
			tenant:     "",
			usageUser:  "alice$t1",
			bucketName: "b1",
		},
		{
			name:       "owner split from tenant field",
			owner:      "alice",
			tenant:     "t1",
			usageUser:  "alice$t1",
			bucketName: "b1",
		},
		{
			name:       "owner has inline tenant plus explicit tenant",
			owner:      "alice$t1",
			tenant:     "t1",
			usageUser:  "alice$t1",
			bucketName: "b2",
		},
		{
			name:       "no tenant",
			owner:      "alice",
			tenant:     "",
			usageUser:  "alice",
			bucketName: "b3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucketUser, bucketTenant := NormalizeUserTenant(tt.owner, tt.tenant)
			bucketKey := BuildUserTenantBucketKey(bucketUser, bucketTenant, tt.bucketName)

			usageUser, usageTenant := NormalizeUserTenant(tt.usageUser, "")
			usageKey := BuildUserTenantBucketKey(usageUser, usageTenant, tt.bucketName)

			if bucketKey != usageKey {
				t.Fatalf("bucket/usage key mismatch: bucketKey=%q usageKey=%q", bucketKey, usageKey)
			}
		})
	}
}

func TestEncodeComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "aGVsbG8=",
		},
		{
			name:     "string with special chars",
			input:    "user$tenant",
			expected: "dXNlciR0ZW5hbnQ=",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: "aGVsbG8gd29ybGQ=",
		},
		{
			name:     "unicode string",
			input:    "日本語",
			expected: "5pel5pys6Kqe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeComponent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeComponent(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "simple string",
			input:       "aGVsbG8=",
			expected:    "hello",
			expectError: false,
		},
		{
			name:        "string with special chars",
			input:       "dXNlciR0ZW5hbnQ=",
			expected:    "user$tenant",
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    "",
			expectError: false,
		},
		{
			name:        "invalid base64",
			input:       "not-valid-base64!!!",
			expected:    "",
			expectError: true,
		},
		{
			name:        "unicode string",
			input:       "5pel5pys6Kqe",
			expected:    "日本語",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeComponent(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	testStrings := []string{
		"hello",
		"user$tenant",
		"user.with.dots",
		"special/chars+test",
		"日本語",
		"",
		"a",
		"long-string-with-many-characters-1234567890",
	}

	for _, s := range testStrings {
		t.Run(s, func(t *testing.T) {
			encoded := EncodeComponent(s)
			decoded, err := DecodeComponent(encoded)
			assert.NoError(t, err)
			assert.Equal(t, s, decoded)
		})
	}
}

func TestSplitUserTenant(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantUser   string
		wantTenant string
	}{
		{
			name:       "no tenant",
			input:      "alice",
			wantUser:   "alice",
			wantTenant: "",
		},
		{
			name:       "user with tenant",
			input:      "alice$tenant1",
			wantUser:   "alice",
			wantTenant: "tenant1",
		},
		{
			name:       "multiple dollar signs",
			input:      "alice$tenant$extra",
			wantUser:   "alice",
			wantTenant: "tenant$extra",
		},
		{
			name:       "empty user with tenant",
			input:      "$tenant",
			wantUser:   "",
			wantTenant: "tenant",
		},
		{
			name:       "user with empty tenant",
			input:      "alice$",
			wantUser:   "alice",
			wantTenant: "",
		},
		{
			name:       "just dollar sign",
			input:      "$",
			wantUser:   "",
			wantTenant: "",
		},
		{
			name:       "empty string",
			input:      "",
			wantUser:   "",
			wantTenant: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, gotTenant := SplitUserTenant(tt.input)
			assert.Equal(t, tt.wantUser, gotUser)
			assert.Equal(t, tt.wantTenant, gotTenant)
		})
	}
}

func TestBuildUserTenantKey(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		tenant   string
		expected string
	}{
		{
			name:     "user and tenant",
			user:     "alice",
			tenant:   "tenant1",
			expected: "YWxpY2U=.dGVuYW50MQ==",
		},
		{
			name:     "user without tenant",
			user:     "alice",
			tenant:   "",
			expected: "YWxpY2U=.none",
		},
		{
			name:     "empty user with tenant",
			user:     "",
			tenant:   "tenant1",
			expected: "none.dGVuYW50MQ==",
		},
		{
			name:     "both empty",
			user:     "",
			tenant:   "",
			expected: "none.none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildUserTenantKey(tt.user, tt.tenant)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildUserTenantBucketKey(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		tenant   string
		bucket   string
		expected string
	}{
		{
			name:     "all fields",
			user:     "alice",
			tenant:   "tenant1",
			bucket:   "mybucket",
			expected: "YWxpY2U=.dGVuYW50MQ==.bXlidWNrZXQ=",
		},
		{
			name:     "no tenant",
			user:     "alice",
			tenant:   "",
			bucket:   "mybucket",
			expected: "YWxpY2U=.none.bXlidWNrZXQ=",
		},
		{
			name:     "no user",
			user:     "",
			tenant:   "tenant1",
			bucket:   "mybucket",
			expected: "none.dGVuYW50MQ==.bXlidWNrZXQ=",
		},
		{
			name:     "only bucket",
			user:     "",
			tenant:   "",
			bucket:   "mybucket",
			expected: "none.none.bXlidWNrZXQ=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildUserTenantBucketKey(tt.user, tt.tenant, tt.bucket)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseKVKey(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		wantUser    string
		wantTenant  string
		wantBucket  string
		expectError bool
	}{
		{
			name:        "user only",
			key:         "YWxpY2U=",
			wantUser:    "alice",
			wantTenant:  "",
			wantBucket:  "",
			expectError: false,
		},
		{
			name:        "user and tenant",
			key:         "YWxpY2U=.dGVuYW50MQ==",
			wantUser:    "alice",
			wantTenant:  "tenant1",
			wantBucket:  "",
			expectError: false,
		},
		{
			name:        "user tenant and bucket",
			key:         "YWxpY2U=.dGVuYW50MQ==.bXlidWNrZXQ=",
			wantUser:    "alice",
			wantTenant:  "tenant1",
			wantBucket:  "mybucket",
			expectError: false,
		},
		{
			name:        "dollar separated key",
			key:         "YWxpY2U=$dGVuYW50MQ==",
			wantUser:    "alice",
			wantTenant:  "tenant1",
			wantBucket:  "",
			expectError: false,
		},
		{
			name:        "too many parts",
			key:         "a.b.c.d",
			wantUser:    "",
			wantTenant:  "",
			wantBucket:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, tenant, bucket, err := ParseKVKey(tt.key)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, user)
				assert.Equal(t, tt.wantTenant, tenant)
				assert.Equal(t, tt.wantBucket, bucket)
			}
		})
	}
}

func TestParseKVKey_InvalidBase64(t *testing.T) {
	// Test with invalid base64 in various positions
	tests := []struct {
		name string
		key  string
	}{
		{
			name: "invalid user",
			key:  "!!!invalid!!!.dGVuYW50MQ==",
		},
		{
			name: "invalid tenant (not placeholder)",
			key:  "YWxpY2U=.!!!invalid!!!",
		},
		{
			name: "invalid bucket",
			key:  "YWxpY2U=.dGVuYW50MQ==.!!!invalid!!!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := ParseKVKey(tt.key)
			assert.Error(t, err)
		})
	}
}

func TestBuildAndParseRoundTrip(t *testing.T) {
	tests := []struct {
		user   string
		tenant string
		bucket string
	}{
		{"alice", "tenant1", "mybucket"},
		{"alice", "tenant1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.user+"/"+tt.tenant+"/"+tt.bucket, func(t *testing.T) {
			if tt.bucket != "" {
				key := BuildUserTenantBucketKey(tt.user, tt.tenant, tt.bucket)
				parsedUser, parsedTenant, parsedBucket, err := ParseKVKey(key)
				assert.NoError(t, err)
				assert.Equal(t, tt.user, parsedUser)
				assert.Equal(t, tt.tenant, parsedTenant)
				assert.Equal(t, tt.bucket, parsedBucket)
			} else if tt.user != "" && tt.tenant != "" {
				key := BuildUserTenantKey(tt.user, tt.tenant)
				parsedUser, parsedTenant, _, err := ParseKVKey(key)
				assert.NoError(t, err)
				assert.Equal(t, tt.user, parsedUser)
				assert.Equal(t, tt.tenant, parsedTenant)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "none", MissingUserPlaceholder)
	assert.Equal(t, "none", MissingTenantPlaceholder)
}
