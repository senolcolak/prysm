// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package opslog

import (
	"testing"

	"github.com/sapcc/go-api-declarations/cadf"
	"github.com/stretchr/testify/assert"
)

func TestMapOperationToAction(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		expected  cadf.Action
	}{
		// Read operations
		{
			name:      "list_buckets",
			operation: "list_buckets",
			expected:  "read/list",
		},
		{
			name:      "list_bucket",
			operation: "list_bucket",
			expected:  "read/list",
		},
		{
			name:      "get_obj",
			operation: "get_obj",
			expected:  "read",
		},
		{
			name:      "get_bucket_info",
			operation: "get_bucket_info",
			expected:  "read",
		},
		{
			name:      "head_obj",
			operation: "head_obj",
			expected:  "read",
		},
		{
			name:      "head_bucket",
			operation: "head_bucket",
			expected:  "read",
		},
		// Create operations
		{
			name:      "put_obj",
			operation: "put_obj",
			expected:  "create",
		},
		{
			name:      "create_bucket",
			operation: "create_bucket",
			expected:  "create",
		},
		// Delete operations
		{
			name:      "delete_obj",
			operation: "delete_obj",
			expected:  "delete",
		},
		{
			name:      "delete_bucket",
			operation: "delete_bucket",
			expected:  "delete",
		},
		// Update operations
		{
			name:      "copy_obj",
			operation: "copy_obj",
			expected:  "update/copy",
		},
		{
			name:      "post_obj",
			operation: "post_obj",
			expected:  "update",
		},
		// Unknown operation
		{
			name:      "unknown_operation",
			operation: "some_random_operation",
			expected:  cadf.UnknownAction,
		},
		{
			name:      "empty_operation",
			operation: "",
			expected:  cadf.UnknownAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapOperationToAction(tt.operation)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractProjectID(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		expected string
	}{
		{
			name:     "user with tenant",
			user:     "projectID$userName",
			expected: "projectID",
		},
		{
			name:     "user without tenant",
			user:     "simpleUser",
			expected: "simpleUser",
		},
		{
			name:     "multiple dollar signs",
			user:     "project$user$extra",
			expected: "project",
		},
		{
			name:     "empty user",
			user:     "",
			expected: "",
		},
		{
			name:     "user starting with dollar",
			user:     "$anonymous",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractProjectID(tt.user)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildHTTPRequest(t *testing.T) {
	tests := []struct {
		name           string
		opLog          *S3OperationLog
		expectedMethod string
		expectedPath   string
	}{
		{
			name: "GET request",
			opLog: &S3OperationLog{
				URI:        "GET /bucket/object.txt HTTP/1.1",
				UserAgent:  "aws-sdk-go/1.0",
				RemoteAddr: "192.168.1.100",
			},
			expectedMethod: "GET",
			expectedPath:   "/bucket/object.txt",
		},
		{
			name: "PUT request",
			opLog: &S3OperationLog{
				URI:        "PUT /bucket/newfile.txt HTTP/1.1",
				UserAgent:  "s3cmd/2.0",
				RemoteAddr: "10.0.0.1",
			},
			expectedMethod: "PUT",
			expectedPath:   "/bucket/newfile.txt",
		},
		{
			name: "DELETE request",
			opLog: &S3OperationLog{
				URI:        "DELETE /bucket/file.txt HTTP/1.1",
				UserAgent:  "curl/7.68.0",
				RemoteAddr: "172.16.0.1",
			},
			expectedMethod: "DELETE",
			expectedPath:   "/bucket/file.txt",
		},
		{
			name: "request with referrer",
			opLog: &S3OperationLog{
				URI:        "GET /bucket HTTP/1.1",
				UserAgent:  "browser/1.0",
				Referrer:   "https://example.com/page",
				RemoteAddr: "192.168.1.50",
			},
			expectedMethod: "GET",
			expectedPath:   "/bucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := buildHTTPRequest(tt.opLog)
			assert.NoError(t, err)
			assert.NotNil(t, req)
			assert.Equal(t, tt.expectedMethod, req.Method)
			assert.Equal(t, tt.expectedPath, req.URL.Path)
			assert.Equal(t, tt.opLog.UserAgent, req.Header.Get("User-Agent"))
			assert.Equal(t, tt.opLog.RemoteAddr, req.RemoteAddr)
		})
	}
}

func TestBuildTarget(t *testing.T) {
	tests := []struct {
		name           string
		opLog          *S3OperationLog
		expectedType   string
		expectedTarget interface{}
	}{
		{
			name: "object target",
			opLog: &S3OperationLog{
				Bucket: "my-bucket",
				Object: "path/to/file.txt",
				User:   "user123",
			},
			expectedType: "*opslog.ObjectTarget",
		},
		{
			name: "bucket target",
			opLog: &S3OperationLog{
				Bucket: "my-bucket",
				Object: "",
				User:   "user123",
			},
			expectedType: "*opslog.BucketTarget",
		},
		{
			name: "account target",
			opLog: &S3OperationLog{
				Bucket: "",
				Object: "",
				User:   "project123$user456",
			},
			expectedType: "*opslog.AccountTarget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := buildTarget(tt.opLog)
			assert.NotNil(t, target)

			switch tt.expectedType {
			case "*opslog.ObjectTarget":
				objTarget, ok := target.(*ObjectTarget)
				assert.True(t, ok)
				assert.Equal(t, tt.opLog.Bucket, objTarget.Bucket)
				assert.Equal(t, tt.opLog.Object, objTarget.Object)
			case "*opslog.BucketTarget":
				bucketTarget, ok := target.(*BucketTarget)
				assert.True(t, ok)
				assert.Equal(t, tt.opLog.Bucket, bucketTarget.Bucket)
			case "*opslog.AccountTarget":
				accountTarget, ok := target.(*AccountTarget)
				assert.True(t, ok)
				assert.Equal(t, "project123", accountTarget.ProjectID)
			}
		})
	}
}

func TestObjectTarget_Render(t *testing.T) {
	target := &ObjectTarget{
		Bucket: "my-bucket",
		Object: "path/to/file.txt",
	}

	resource := target.Render()

	assert.Equal(t, "service/storage/object", string(resource.TypeURI))
	assert.Equal(t, "my-bucket/path/to/file.txt", resource.ID)
	assert.Equal(t, "path/to/file.txt", resource.Name)
	assert.Len(t, resource.Attachments, 1)
	assert.Equal(t, "bucket", resource.Attachments[0].Name)
	assert.Equal(t, "my-bucket", resource.Attachments[0].Content)
}

func TestBucketTarget_Render(t *testing.T) {
	target := &BucketTarget{
		Bucket: "my-bucket",
	}

	resource := target.Render()

	assert.Equal(t, "service/storage/bucket", string(resource.TypeURI))
	assert.Equal(t, "my-bucket", resource.ID)
	assert.Equal(t, "my-bucket", resource.Name)
}

func TestAccountTarget_Render(t *testing.T) {
	target := &AccountTarget{
		ProjectID: "project123",
	}

	resource := target.Render()

	assert.Equal(t, "service/storage/account", string(resource.TypeURI))
	assert.Equal(t, "project123", resource.ID)
	assert.Equal(t, "project123", resource.Name)
}

func TestBuildUserInfo_WithKeystoneScope(t *testing.T) {
	opLog := &S3OperationLog{
		User: "user123$tenant456",
		KeystoneScope: &KeystoneScope{
			Project: KeystoneProject{
				ID:   "proj-id",
				Name: "proj-name",
				Domain: KeystoneDomain{
					ID:   "domain-id",
					Name: "domain-name",
				},
			},
			User: KeystoneUser{
				ID:   "user-id",
				Name: "user-name",
				Domain: KeystoneDomain{
					ID:   "user-domain-id",
					Name: "user-domain-name",
				},
			},
			Roles: []string{"admin", "member"},
			ApplicationCredential: &KeystoneApplicationCredential{
				ID:   "app-cred-id",
				Name: "app-cred-name",
			},
		},
	}

	userInfo := buildUserInfo(opLog)
	assert.NotNil(t, userInfo)

	keystoneUser, ok := userInfo.(*KeystoneUserInfo)
	assert.True(t, ok)
	assert.Equal(t, "proj-id", keystoneUser.ProjectID)
	assert.Equal(t, "proj-name", keystoneUser.ProjectName)
	assert.Equal(t, "domain-id", keystoneUser.DomainID)
	assert.Equal(t, "domain-name", keystoneUser.DomainName)
	assert.Equal(t, "user-id", keystoneUser.UserID)
	assert.Equal(t, "user-name", keystoneUser.UserName)
	assert.Equal(t, "user-domain-name", keystoneUser.UserDomain)
	assert.Equal(t, []string{"admin", "member"}, keystoneUser.Roles)
	assert.Equal(t, "app-cred-id", keystoneUser.AppCredID)
	assert.Equal(t, "app-cred-name", keystoneUser.AppCredName)
}

func TestBuildUserInfo_WithoutKeystoneScope(t *testing.T) {
	opLog := &S3OperationLog{
		User:          "project123$user456",
		KeystoneScope: nil,
	}

	userInfo := buildUserInfo(opLog)
	assert.NotNil(t, userInfo)

	simpleUser, ok := userInfo.(*SimpleUserInfo)
	assert.True(t, ok)
	assert.Equal(t, "project123$user456", simpleUser.UserID)
	assert.Equal(t, "project123", simpleUser.ProjectID)
}

func TestGetAppCredID(t *testing.T) {
	tests := []struct {
		name     string
		appCred  *KeystoneApplicationCredential
		expected string
	}{
		{
			name:     "nil app credential",
			appCred:  nil,
			expected: "",
		},
		{
			name: "with app credential",
			appCred: &KeystoneApplicationCredential{
				ID:   "cred-123",
				Name: "my-credential",
			},
			expected: "cred-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAppCredID(tt.appCred)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAppCredName(t *testing.T) {
	tests := []struct {
		name     string
		appCred  *KeystoneApplicationCredential
		expected string
	}{
		{
			name:     "nil app credential",
			appCred:  nil,
			expected: "",
		},
		{
			name: "with app credential",
			appCred: &KeystoneApplicationCredential{
				ID:   "cred-123",
				Name: "my-credential",
			},
			expected: "my-credential",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAppCredName(tt.appCred)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKeystoneUserInfo_AsInitiator(t *testing.T) {
	userInfo := &KeystoneUserInfo{
		ProjectID:   "proj-123",
		ProjectName: "my-project",
		DomainID:    "domain-123",
		DomainName:  "my-domain",
		UserID:      "user-123",
		UserName:    "john.doe",
		UserDomain:  "user-domain",
		Roles:       []string{"admin"},
		AppCredID:   "app-123",
		AppCredName: "my-app-cred",
	}

	host := cadf.Host{
		Address: "192.168.1.100",
	}

	initiator := userInfo.AsInitiator(host)

	assert.Equal(t, "service/security/account/user", string(initiator.TypeURI))
	assert.Equal(t, "user-123", initiator.ID)
	assert.Equal(t, "john.doe", initiator.Name)
	assert.Equal(t, "proj-123", initiator.ProjectID)
	assert.Equal(t, "my-project", initiator.ProjectName)
	assert.Equal(t, "domain-123", initiator.DomainID)
	assert.Equal(t, "my-domain", initiator.DomainName)
	assert.Equal(t, "app-123", initiator.AppCredentialID)
}

func TestSimpleUserInfo_AsInitiator(t *testing.T) {
	userInfo := &SimpleUserInfo{
		UserID:    "user-123",
		ProjectID: "proj-123",
	}

	host := cadf.Host{
		Address: "10.0.0.1",
	}

	initiator := userInfo.AsInitiator(host)

	assert.Equal(t, "service/security/account/user", string(initiator.TypeURI))
	assert.Equal(t, "user-123", initiator.ID)
	assert.Equal(t, "user-123", initiator.Name)
	assert.Equal(t, "proj-123", initiator.ProjectID)
}

func TestS3OperationLog_ToAuditEvent(t *testing.T) {
	opLog := &S3OperationLog{
		Bucket:     "my-bucket",
		Object:     "file.txt",
		Time:       "2024-01-15T12:30:45.123456Z",
		RemoteAddr: "192.168.1.100",
		User:       "project$user",
		Operation:  "get_obj",
		URI:        "GET /my-bucket/file.txt HTTP/1.1",
		HTTPStatus: "200",
		UserAgent:  "aws-cli/2.0",
	}

	event, err := opLog.ToAuditEvent()
	assert.NoError(t, err)
	assert.Equal(t, "read", string(event.Action))
	assert.Equal(t, 200, event.ReasonCode)
}

func TestS3OperationLog_ToAuditEvent_InvalidTimestamp(t *testing.T) {
	opLog := &S3OperationLog{
		Bucket:     "my-bucket",
		Time:       "invalid-timestamp",
		User:       "user",
		Operation:  "get_obj",
		URI:        "GET /my-bucket HTTP/1.1",
		HTTPStatus: "200",
	}

	event, err := opLog.ToAuditEvent()
	assert.NoError(t, err) // Should not error, just use current time
	assert.NotZero(t, event.Time)
}

func TestS3OperationLog_ToAuditEvent_InvalidHTTPStatus(t *testing.T) {
	opLog := &S3OperationLog{
		Bucket:     "my-bucket",
		Time:       "2024-01-15T12:30:45.123456Z",
		User:       "user",
		Operation:  "get_obj",
		URI:        "GET /my-bucket HTTP/1.1",
		HTTPStatus: "invalid",
	}

	event, err := opLog.ToAuditEvent()
	assert.NoError(t, err)
	assert.Equal(t, 500, event.ReasonCode) // Defaults to 500
}
