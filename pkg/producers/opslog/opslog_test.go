// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package opslog

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS3OperationLog_CleanupBucketName_Extended(t *testing.T) {
	tests := []struct {
		name     string
		bucket   string
		expected string
	}{
		{
			name:     "simple bucket name",
			bucket:   "my-bucket",
			expected: "my-bucket",
		},
		{
			name:     "bucket with tenant prefix",
			bucket:   "tenant/my-bucket",
			expected: "my-bucket",
		},
		{
			name:     "bucket with multiple slashes",
			bucket:   "tenant/user/my-bucket",
			expected: "my-bucket",
		},
		{
			name:     "bucket with deep path",
			bucket:   "a/b/c/d/my-bucket",
			expected: "my-bucket",
		},
		{
			name:     "empty bucket",
			bucket:   "",
			expected: "",
		},
		{
			name:     "single slash",
			bucket:   "/",
			expected: "",
		},
		{
			name:     "trailing slash",
			bucket:   "bucket/",
			expected: "",
		},
		{
			name:     "leading slash",
			bucket:   "/bucket",
			expected: "bucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &S3OperationLog{Bucket: tt.bucket}
			log.CleanupBucketName()
			assert.Equal(t, tt.expected, log.Bucket)
		})
	}
}

func TestExtractUserAndTenant_Extended(t *testing.T) {
	tests := []struct {
		name           string
		user           string
		expectedUser   string
		expectedTenant string
	}{
		{
			name:           "user with tenant",
			user:           "tenant$user",
			expectedUser:   "tenant",
			expectedTenant: "user",
		},
		{
			name:           "user without tenant",
			user:           "simpleuser",
			expectedUser:   "simpleuser",
			expectedTenant: "none",
		},
		{
			name:           "empty user",
			user:           "",
			expectedUser:   "",
			expectedTenant: "none",
		},
		{
			name:           "multiple dollar signs",
			user:           "a$b$c",
			expectedUser:   "a",
			expectedTenant: "b$c",
		},
		{
			name:           "dollar at start",
			user:           "$tenant",
			expectedUser:   "",
			expectedTenant: "tenant",
		},
		{
			name:           "dollar at end",
			user:           "user$",
			expectedUser:   "user",
			expectedTenant: "",
		},
		{
			name:           "only dollar sign",
			user:           "$",
			expectedUser:   "",
			expectedTenant: "",
		},
		{
			name:           "project and user format",
			user:           "project123$admin",
			expectedUser:   "project123",
			expectedTenant: "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, tenant := extractUserAndTenant(tt.user)
			assert.Equal(t, tt.expectedUser, user)
			assert.Equal(t, tt.expectedTenant, tenant)
		})
	}
}

func TestS3OperationLog_FullJSONRoundTrip(t *testing.T) {
	original := S3OperationLog{
		Bucket:             "test-bucket",
		Object:             "path/to/file.txt",
		Time:               "2024-01-15T12:30:45.123456Z",
		TimeLocal:          "15/Jan/2024:12:30:45 +0000",
		RemoteAddr:         "192.168.1.100",
		User:               "tenant$user",
		Operation:          "get_obj",
		URI:                "GET /test-bucket/path/to/file.txt HTTP/1.1",
		HTTPStatus:         "200",
		ErrorCode:          "",
		BytesSent:          1024,
		BytesReceived:      0,
		ObjectSize:         1024,
		TotalTime:          150,
		UserAgent:          "aws-cli/2.0",
		Referrer:           "https://example.com",
		TransID:            "tx123456",
		AuthenticationType: "keystone",
		AccessKeyID:        "AKIAIOSFODNN7EXAMPLE",
		TempURL:            false,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	assert.NoError(t, err)

	// Unmarshal back
	var restored S3OperationLog
	err = json.Unmarshal(jsonData, &restored)
	assert.NoError(t, err)

	// Verify all fields
	assert.Equal(t, original.Bucket, restored.Bucket)
	assert.Equal(t, original.Object, restored.Object)
	assert.Equal(t, original.Time, restored.Time)
	assert.Equal(t, original.RemoteAddr, restored.RemoteAddr)
	assert.Equal(t, original.User, restored.User)
	assert.Equal(t, original.Operation, restored.Operation)
	assert.Equal(t, original.URI, restored.URI)
	assert.Equal(t, original.HTTPStatus, restored.HTTPStatus)
	assert.Equal(t, original.BytesSent, restored.BytesSent)
	assert.Equal(t, original.TotalTime, restored.TotalTime)
	assert.Equal(t, original.TempURL, restored.TempURL)
}

func TestS3OperationLog_WithKeystoneScope(t *testing.T) {
	log := S3OperationLog{
		Bucket:    "bucket",
		User:      "user",
		Operation: "get_obj",
		KeystoneScope: &KeystoneScope{
			Project: KeystoneProject{
				ID:   "proj-123",
				Name: "my-project",
				Domain: KeystoneDomain{
					ID:   "domain-123",
					Name: "my-domain",
				},
			},
			User: KeystoneUser{
				ID:   "user-123",
				Name: "john.doe",
				Domain: KeystoneDomain{
					ID:   "user-domain-123",
					Name: "user-domain",
				},
			},
			Roles: []string{"admin", "member"},
			ApplicationCredential: &KeystoneApplicationCredential{
				ID:         "app-cred-123",
				Name:       "my-app-cred",
				Restricted: false,
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(log)
	assert.NoError(t, err)

	// Verify KeystoneScope is present
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	keystoneScope, ok := result["keystone_scope"].(map[string]interface{})
	assert.True(t, ok, "keystone_scope should be present")

	project := keystoneScope["project"].(map[string]interface{})
	assert.Equal(t, "proj-123", project["id"])
	assert.Equal(t, "my-project", project["name"])
}

func TestS3OperationLog_WithoutKeystoneScope(t *testing.T) {
	log := S3OperationLog{
		Bucket:        "bucket",
		User:          "user",
		Operation:     "get_obj",
		KeystoneScope: nil,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(log)
	assert.NoError(t, err)

	// Verify KeystoneScope is omitted
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	_, ok := result["keystone_scope"]
	assert.False(t, ok, "keystone_scope should be omitted when nil")
}

func TestKeystoneDomain_StructFields(t *testing.T) {
	domain := KeystoneDomain{
		ID:   "domain-123",
		Name: "my-domain",
	}

	assert.Equal(t, "domain-123", domain.ID)
	assert.Equal(t, "my-domain", domain.Name)

	// JSON serialization
	jsonData, err := json.Marshal(domain)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Equal(t, "domain-123", result["id"])
	assert.Equal(t, "my-domain", result["name"])
}

func TestKeystoneProject_StructFields(t *testing.T) {
	project := KeystoneProject{
		ID:   "proj-123",
		Name: "my-project",
		Domain: KeystoneDomain{
			ID:   "domain-123",
			Name: "my-domain",
		},
	}

	assert.Equal(t, "proj-123", project.ID)
	assert.Equal(t, "my-project", project.Name)
	assert.Equal(t, "domain-123", project.Domain.ID)
	assert.Equal(t, "my-domain", project.Domain.Name)
}

func TestKeystoneUser_StructFields(t *testing.T) {
	user := KeystoneUser{
		ID:   "user-123",
		Name: "john.doe",
		Domain: KeystoneDomain{
			ID:   "user-domain-123",
			Name: "user-domain",
		},
	}

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "john.doe", user.Name)
	assert.Equal(t, "user-domain-123", user.Domain.ID)
}

func TestKeystoneApplicationCredential_StructFields(t *testing.T) {
	appCred := KeystoneApplicationCredential{
		ID:         "app-cred-123",
		Name:       "my-app-cred",
		Restricted: true,
	}

	assert.Equal(t, "app-cred-123", appCred.ID)
	assert.Equal(t, "my-app-cred", appCred.Name)
	assert.True(t, appCred.Restricted)
}

func TestKeystoneScope_JSONSerialization(t *testing.T) {
	scope := KeystoneScope{
		Project: KeystoneProject{
			ID:   "proj-123",
			Name: "my-project",
			Domain: KeystoneDomain{
				ID:   "domain-123",
				Name: "my-domain",
			},
		},
		User: KeystoneUser{
			ID:   "user-123",
			Name: "john.doe",
			Domain: KeystoneDomain{
				ID:   "user-domain-123",
				Name: "user-domain",
			},
		},
		Roles: []string{"admin", "member", "reader"},
		ApplicationCredential: &KeystoneApplicationCredential{
			ID:         "app-cred-123",
			Name:       "my-app-cred",
			Restricted: false,
		},
	}

	// Marshal
	jsonData, err := json.Marshal(scope)
	assert.NoError(t, err)

	// Unmarshal
	var restored KeystoneScope
	err = json.Unmarshal(jsonData, &restored)
	assert.NoError(t, err)

	assert.Equal(t, scope.Project.ID, restored.Project.ID)
	assert.Equal(t, scope.User.Name, restored.User.Name)
	assert.Equal(t, scope.Roles, restored.Roles)
	assert.Equal(t, scope.ApplicationCredential.ID, restored.ApplicationCredential.ID)
}

func TestKeystoneScope_WithoutApplicationCredential(t *testing.T) {
	scope := KeystoneScope{
		Project: KeystoneProject{
			ID:   "proj-123",
			Name: "my-project",
		},
		User: KeystoneUser{
			ID:   "user-123",
			Name: "john.doe",
		},
		Roles:                 []string{"member"},
		ApplicationCredential: nil,
	}

	jsonData, err := json.Marshal(scope)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// application_credential should be omitted when nil
	_, ok := result["application_credential"]
	assert.False(t, ok, "application_credential should be omitted when nil")
}

func TestS3OperationLog_DefaultValues(t *testing.T) {
	log := S3OperationLog{}

	assert.Empty(t, log.Bucket)
	assert.Empty(t, log.Object)
	assert.Empty(t, log.User)
	assert.Empty(t, log.Operation)
	assert.Equal(t, 0, log.BytesSent)
	assert.Equal(t, 0, log.BytesReceived)
	assert.Equal(t, 0, log.TotalTime)
	assert.False(t, log.TempURL)
	assert.Nil(t, log.KeystoneScope)
}

func TestS3OperationLog_ParseFromRealJSON(t *testing.T) {
	// Simulate a real RGW ops log entry
	jsonData := `{
		"bucket": "my-bucket",
		"object": "path/to/file.txt",
		"time": "2024-01-15T12:30:45.123456Z",
		"time_local": "15/Jan/2024:12:30:45 +0000",
		"remote_addr": "10.0.0.1",
		"user": "project123$admin",
		"operation": "put_obj",
		"uri": "PUT /my-bucket/path/to/file.txt HTTP/1.1",
		"http_status": "201",
		"error_code": "",
		"bytes_sent": 0,
		"bytes_received": 2048,
		"object_size": 2048,
		"total_time": 250,
		"user_agent": "aws-sdk-go/1.0",
		"referrer": "",
		"trans_id": "tx987654",
		"authentication_type": "s3",
		"access_key_id": "AKIAEXAMPLE",
		"temp_url": false
	}`

	var log S3OperationLog
	err := json.Unmarshal([]byte(jsonData), &log)
	assert.NoError(t, err)

	assert.Equal(t, "my-bucket", log.Bucket)
	assert.Equal(t, "path/to/file.txt", log.Object)
	assert.Equal(t, "project123$admin", log.User)
	assert.Equal(t, "put_obj", log.Operation)
	assert.Equal(t, "201", log.HTTPStatus)
	assert.Equal(t, 2048, log.BytesReceived)
	assert.Equal(t, 250, log.TotalTime)
	assert.Equal(t, "s3", log.AuthenticationType)
}
