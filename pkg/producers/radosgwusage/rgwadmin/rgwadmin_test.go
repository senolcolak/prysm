// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package rgwadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildQueryPath(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		path     string
		args     string
		expected string
	}{
		{
			name:     "simple path",
			endpoint: "http://localhost:8080",
			path:     "/user",
			args:     "uid=admin&format=json",
			expected: "http://localhost:8080/admin/user?uid=admin&format=json",
		},
		{
			name:     "path with existing query",
			endpoint: "http://localhost:8080",
			path:     "/user?stats=true",
			args:     "uid=admin",
			expected: "http://localhost:8080/admin/user?stats=true&uid=admin",
		},
		{
			name:     "empty args",
			endpoint: "http://localhost:8080",
			path:     "/bucket",
			args:     "",
			expected: "http://localhost:8080/admin/bucket?",
		},
		{
			name:     "metadata path",
			endpoint: "https://rgw.example.com",
			path:     "/metadata/user",
			args:     "format=json",
			expected: "https://rgw.example.com/admin/metadata/user?format=json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildQueryPath(tt.endpoint, tt.path, tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorReason_Error(t *testing.T) {
	tests := []struct {
		name     string
		reason   errorReason
		expected string
	}{
		{name: "NoSuchUser", reason: ErrNoSuchUser, expected: "NoSuchUser"},
		{name: "AccessDenied", reason: ErrAccessDenied, expected: "AccessDenied"},
		{name: "UserExists", reason: ErrUserExists, expected: "UserAlreadyExists"},
		{name: "Unknown", reason: ErrUnknown, expected: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.reason.Error())
		})
	}
}

func TestStatusError_Error(t *testing.T) {
	err := statusError{
		Code:      "NoSuchUser",
		RequestID: "req-123",
		HostID:    "host-456",
	}

	errStr := err.Error()
	assert.Contains(t, errStr, "NoSuchUser")
	assert.Contains(t, errStr, "req-123")
	assert.Contains(t, errStr, "host-456")
}

func TestStatusError_Is(t *testing.T) {
	t.Run("matches known error reason", func(t *testing.T) {
		err := statusError{Code: "NoSuchUser"}
		assert.True(t, errors.Is(err, ErrNoSuchUser))
	})

	t.Run("does not match different error reason", func(t *testing.T) {
		err := statusError{Code: "NoSuchUser"}
		assert.False(t, errors.Is(err, ErrAccessDenied))
	})

	t.Run("does not match non-errorReason", func(t *testing.T) {
		err := statusError{Code: "NoSuchUser"}
		assert.False(t, errors.Is(err, errors.New("some error")))
	})
}

func TestHandleStatusError(t *testing.T) {
	t.Run("valid error response", func(t *testing.T) {
		resp := `{"Code":"NoSuchUser","RequestId":"req-123","HostId":"host-456"}`
		err := handleStatusError([]byte(resp))
		assert.Error(t, err)

		var se statusError
		assert.True(t, errors.As(err, &se))
		assert.Equal(t, "NoSuchUser", se.Code)
	})

	t.Run("empty code returns unknown error", func(t *testing.T) {
		resp := `{"Code":"","RequestId":"req-123"}`
		err := handleStatusError([]byte(resp))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown error")
	})

	t.Run("invalid JSON returns unmarshal error", func(t *testing.T) {
		err := handleStatusError([]byte("not json"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), unmarshalError)
	})
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		accessKey string
		secretKey string
		wantErr   error
	}{
		{
			name:      "valid config",
			endpoint:  "http://localhost:8080",
			accessKey: "access",
			secretKey: "secret",
			wantErr:   nil,
		},
		{
			name:      "missing endpoint",
			endpoint:  "",
			accessKey: "access",
			secretKey: "secret",
			wantErr:   errNoEndpoint,
		},
		{
			name:      "missing access key",
			endpoint:  "http://localhost:8080",
			accessKey: "",
			secretKey: "secret",
			wantErr:   errNoAccessKey,
		},
		{
			name:      "missing secret key",
			endpoint:  "http://localhost:8080",
			accessKey: "access",
			secretKey: "",
			wantErr:   errNoSecretKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.endpoint, tt.accessKey, tt.secretKey)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("valid creation", func(t *testing.T) {
		api, err := New("http://localhost:8080", "access", "secret", nil)
		assert.NoError(t, err)
		assert.NotNil(t, api)
		assert.Equal(t, "http://localhost:8080", api.Endpoint)
		assert.Equal(t, "access", api.Auth.AccessKey)
		assert.Equal(t, "secret", api.Auth.SecretKey)
		assert.NotNil(t, api.HTTPClient)
	})

	t.Run("missing endpoint", func(t *testing.T) {
		api, err := New("", "access", "secret", nil)
		assert.Error(t, err)
		assert.Nil(t, api)
	})

	t.Run("missing access key", func(t *testing.T) {
		api, err := New("http://localhost:8080", "", "secret", nil)
		assert.Error(t, err)
		assert.Nil(t, api)
	})
}

func TestUser_GetKVUser(t *testing.T) {
	suspended := 0
	maxBuckets := 100

	user := User{
		ID:                  "user1",
		DisplayName:         "Test User",
		Email:               "test@example.com",
		Suspended:           &suspended,
		MaxBuckets:          &maxBuckets,
		OpMask:              "read, write",
		DefaultPlacement:    "default",
		DefaultStorageClass: "STANDARD",
		Type:                "rgw",
		Tenant:              "tenant1",
		Caps: []UserCapSpec{
			{Type: "users", Perm: "*"},
		},
		BucketQuota: QuotaSpec{
			Enabled: new(bool),
		},
	}

	kvUser := user.GetKVUser()
	assert.NotNil(t, kvUser)
	assert.Equal(t, "user1", kvUser.ID)
	assert.Equal(t, "Test User", kvUser.DisplayName)
	assert.Equal(t, "test@example.com", kvUser.Email)
	assert.Equal(t, &suspended, kvUser.Suspended)
	assert.Equal(t, &maxBuckets, kvUser.MaxBuckets)
	assert.Equal(t, "tenant1", kvUser.Tenant)
	assert.Len(t, kvUser.Caps, 1)
}

func TestKVUser_GetUserIdentification(t *testing.T) {
	tests := []struct {
		name     string
		user     KVUser
		expected string
	}{
		{
			name:     "user with tenant",
			user:     KVUser{ID: "user1", Tenant: "tenant1"},
			expected: "user1$tenant1",
		},
		{
			name:     "user without tenant",
			user:     KVUser{ID: "user1", Tenant: ""},
			expected: "user1",
		},
		{
			name:     "empty user with tenant",
			user:     KVUser{ID: "", Tenant: "tenant1"},
			expected: "$tenant1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetUserIdentification()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateUserRequest(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name:    "valid user with ID",
			user:    User{ID: "user1"},
			wantErr: false,
		},
		{
			name: "valid user with access key",
			user: User{
				Keys: []UserKeySpec{{AccessKey: "key1"}},
			},
			wantErr: false,
		},
		{
			name:    "missing ID and no keys",
			user:    User{},
			wantErr: true,
		},
		{
			name: "empty access key in keys",
			user: User{
				Keys: []UserKeySpec{{AccessKey: ""}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserRequest(tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"user_id": "testuser",
		"display_name": "Test User",
		"email": "test@example.com",
		"suspended": 0,
		"max_buckets": 1000,
		"keys": [
			{
				"user": "testuser",
				"access_key": "AKIAIOSFODNN7EXAMPLE",
				"secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
			}
		],
		"caps": [
			{"type": "users", "perm": "*"},
			{"type": "buckets", "perm": "read"}
		],
		"type": "rgw",
		"tenant": "mycompany"
	}`

	var user User
	err := json.Unmarshal([]byte(jsonData), &user)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.ID)
	assert.Equal(t, "Test User", user.DisplayName)
	assert.Len(t, user.Keys, 1)
	assert.Len(t, user.Caps, 2)
	assert.Equal(t, "mycompany", user.Tenant)
}

func TestValueToURLParams(t *testing.T) {
	type testStruct struct {
		Name  string `url:"name"`
		Value int    `url:"value"`
		Skip  string `url:"-"`
	}

	s := testStruct{Name: "test", Value: 42, Skip: "ignored"}
	params := valueToURLParams(s, []string{"name", "value"})

	assert.Equal(t, "json", params.Get("format"))
	assert.Equal(t, "test", params.Get("name"))
	assert.Equal(t, "42", params.Get("value"))
	assert.Empty(t, params.Get("skip"))
}

func TestValueToURLParams_FiltersFields(t *testing.T) {
	type testStruct struct {
		Name  string `url:"name"`
		Value int    `url:"value"`
	}

	s := testStruct{Name: "test", Value: 42}
	params := valueToURLParams(s, []string{"name"}) // Only allow "name"

	assert.Equal(t, "test", params.Get("name"))
	assert.Empty(t, params.Get("value")) // Should be filtered out
}

func TestValueToURLParams_EmptyString(t *testing.T) {
	type testStruct struct {
		Name string `url:"name"`
	}

	s := testStruct{Name: ""}
	params := valueToURLParams(s, []string{"name"})

	assert.Empty(t, params.Get("name")) // Empty strings should be skipped
}

func TestKVUser_JSONRoundTrip(t *testing.T) {
	original := KVUser{
		ID:          "user1",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Tenant:      "tenant1",
		Type:        "rgw",
	}

	jsonData, err := json.Marshal(original)
	assert.NoError(t, err)

	var restored KVUser
	err = json.Unmarshal(jsonData, &restored)
	assert.NoError(t, err)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.DisplayName, restored.DisplayName)
	assert.Equal(t, original.Email, restored.Email)
	assert.Equal(t, original.Tenant, restored.Tenant)
}

// === Mock HTTP Client for API tests ===

type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func newMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

func TestParseResponse_Success(t *testing.T) {
	resp := newMockResponse(200, `{"result":"ok"}`)
	body, err := parseResponse(resp)
	assert.NoError(t, err)
	assert.Equal(t, `{"result":"ok"}`, string(body))
}

func TestParseResponse_ErrorStatus(t *testing.T) {
	resp := newMockResponse(404, `{"Code":"NoSuchUser","RequestId":"req-1","HostId":"host-1"}`)
	_, err := parseResponse(resp)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoSuchUser))
}

func TestParseResponse_ServerError(t *testing.T) {
	resp := newMockResponse(500, `{"Code":"InternalError","RequestId":"req-2","HostId":"host-2"}`)
	_, err := parseResponse(resp)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInternalError))
}

func TestBucket_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"bucket": "test-bucket",
		"num_shards": 16,
		"tenant": "mycompany",
		"zonegroup": "default",
		"placement_rule": "default-placement",
		"id": "bucket-id-123",
		"marker": "marker-123",
		"index_type": "Normal",
		"owner": "admin$admin",
		"ver": "0#1",
		"master_ver": "0#0",
		"mtime": "2024-01-15T12:00:00.000000Z",
		"usage": {
			"rgw.main": {
				"size": 1024,
				"size_actual": 2048,
				"num_objects": 10
			},
			"rgw.multimeta": {
				"size": 0,
				"num_objects": 0
			}
		},
		"bucket_quota": {
			"enabled": false,
			"max_size": -1,
			"max_objects": -1
		}
	}`

	var bucket Bucket
	err := json.Unmarshal([]byte(jsonData), &bucket)
	assert.NoError(t, err)
	assert.Equal(t, "test-bucket", bucket.Bucket)
	assert.Equal(t, "mycompany", bucket.Tenant)
	assert.Equal(t, "admin$admin", bucket.Owner)
	assert.NotNil(t, bucket.NumShards)
	assert.Equal(t, uint64(16), *bucket.NumShards)
	assert.NotNil(t, bucket.Usage.RgwMain.Size)
	assert.Equal(t, uint64(1024), *bucket.Usage.RgwMain.Size)
	assert.NotNil(t, bucket.Usage.RgwMain.NumObjects)
	assert.Equal(t, uint64(10), *bucket.Usage.RgwMain.NumObjects)
}

func TestUsage_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"entries": [
			{
				"user": "admin",
				"buckets": [
					{
						"bucket": "test-bucket",
						"time": "2024-01-15 12:00:00.000000Z",
						"epoch": 1705320000,
						"owner": "admin",
						"categories": [
							{
								"category": "get_obj",
								"bytes_sent": 1024,
								"bytes_received": 0,
								"ops": 10,
								"successful_ops": 10
							},
							{
								"category": "put_obj",
								"bytes_sent": 0,
								"bytes_received": 2048,
								"ops": 5,
								"successful_ops": 5
							}
						]
					}
				]
			}
		],
		"summary": [
			{
				"user": "admin",
				"categories": [
					{
						"category": "get_obj",
						"bytes_sent": 1024,
						"bytes_received": 0,
						"ops": 10,
						"successful_ops": 10
					}
				],
				"total": {
					"bytes_sent": 1024,
					"bytes_received": 2048,
					"ops": 15,
					"successful_ops": 15
				}
			}
		]
	}`

	var usage Usage
	err := json.Unmarshal([]byte(jsonData), &usage)
	assert.NoError(t, err)
	assert.Len(t, usage.Entries, 1)
	assert.Equal(t, "admin", usage.Entries[0].User)
	assert.Len(t, usage.Entries[0].Buckets, 1)
	assert.Len(t, usage.Entries[0].Buckets[0].Categories, 2)
	assert.Equal(t, uint64(1024), usage.Entries[0].Buckets[0].Categories[0].BytesSent)
	assert.Len(t, usage.Summary, 1)
	assert.Equal(t, uint64(15), usage.Summary[0].Total.Ops)
}

func TestQuotaSpec_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"enabled": true,
		"check_on_raw": false,
		"max_size": 1073741824,
		"max_size_kb": 1048576,
		"max_objects": 10000
	}`

	var quota QuotaSpec
	err := json.Unmarshal([]byte(jsonData), &quota)
	assert.NoError(t, err)
	assert.NotNil(t, quota.Enabled)
	assert.True(t, *quota.Enabled)
	assert.NotNil(t, quota.MaxSize)
	assert.Equal(t, int64(1073741824), *quota.MaxSize)
	assert.NotNil(t, quota.MaxObjects)
	assert.Equal(t, int64(10000), *quota.MaxObjects)
}

func TestValueToURLParams_BoolField(t *testing.T) {
	type testStruct struct {
		Active bool `url:"active"`
	}

	s := testStruct{Active: true}
	params := valueToURLParams(s, []string{"active"})
	assert.Equal(t, "true", params.Get("active"))

	s2 := testStruct{Active: false}
	params2 := valueToURLParams(s2, []string{"active"})
	assert.Equal(t, "false", params2.Get("active"))
}

func TestValueToURLParams_PointerField(t *testing.T) {
	val := true
	type testStruct struct {
		Enabled *bool `url:"enabled"`
	}

	s := testStruct{Enabled: &val}
	params := valueToURLParams(s, []string{"enabled"})
	assert.Equal(t, "true", params.Get("enabled"))

	s2 := testStruct{Enabled: nil}
	params2 := valueToURLParams(s2, []string{"enabled"})
	assert.Empty(t, params2.Get("enabled"))
}

func TestValueToURLParams_Int64Field(t *testing.T) {
	type testStruct struct {
		Count int64 `url:"count"`
	}

	s := testStruct{Count: 42}
	params := valueToURLParams(s, []string{"count"})
	assert.Equal(t, "42", params.Get("count"))

	s2 := testStruct{Count: 0}
	params2 := valueToURLParams(s2, []string{"count"})
	assert.Empty(t, params2.Get("count")) // Zero int should be skipped
}

func TestAddToURLParams(t *testing.T) {
	type testStruct struct {
		Name string `url:"name"`
	}

	params := make(map[string][]string)
	urlValues := (*url.Values)(&params)
	urlValues.Add("format", "json")

	s := testStruct{Name: "test"}
	addToURLParams(urlValues, s, []string{"name"})
	assert.Equal(t, "test", urlValues.Get("name"))
	assert.Equal(t, "json", urlValues.Get("format"))
}

func TestPopulateURLParams_NilPointer(t *testing.T) {
	type testStruct struct {
		Name string `url:"name"`
	}

	var nilPtr *testStruct
	allowed := map[string]struct{}{"name": {}}
	values := url.Values{}

	// Should not panic with nil pointer
	populateURLParams(nilPtr, allowed, &values)
	assert.Empty(t, values.Get("name"))
}

func TestErrorReasonConstants(t *testing.T) {
	// Verify error constants are defined correctly
	assert.Equal(t, "UserAlreadyExists", ErrUserExists.Error())
	assert.Equal(t, "NoSuchUser", ErrNoSuchUser.Error())
	assert.Equal(t, "AccessDenied", ErrAccessDenied.Error())
	assert.Equal(t, "NoSuchBucket", ErrNoSuchBucket.Error())
	assert.Equal(t, "InternalError", ErrInternalError.Error())
	assert.Equal(t, "SignatureDoesNotMatch", ErrSignatureDoesNotMatch.Error())
	assert.Equal(t, "InvalidArgument", ErrInvalidArgument.Error())
	assert.Equal(t, "BucketNotEmpty", ErrBucketNotEmpty.Error())
}

func TestUserStat_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"size": 1024,
		"size_rounded": 4096,
		"num_objects": 10
	}`

	var stat UserStat
	err := json.Unmarshal([]byte(jsonData), &stat)
	assert.NoError(t, err)
	assert.NotNil(t, stat.Size)
	assert.Equal(t, uint64(1024), *stat.Size)
	assert.NotNil(t, stat.SizeRounded)
	assert.Equal(t, uint64(4096), *stat.SizeRounded)
	assert.NotNil(t, stat.NumObjects)
	assert.Equal(t, uint64(10), *stat.NumObjects)
}

func TestAPI_GetUsers(t *testing.T) {
	mock := &mockHTTPClient{
		response: newMockResponse(200, `["user1","user2","admin"]`),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	users, err := api.GetUsers(context.Background())
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.Contains(t, users, "user1")
	assert.Contains(t, users, "admin")
}

func TestAPI_GetUsers_Error(t *testing.T) {
	mock := &mockHTTPClient{
		response: newMockResponse(403, `{"Code":"AccessDenied","RequestId":"req-1","HostId":"host-1"}`),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	_, err = api.GetUsers(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrAccessDenied))
}

func TestAPI_GetUsers_HTTPFailure(t *testing.T) {
	mock := &mockHTTPClient{
		err: errors.New("connection refused"),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	_, err = api.GetUsers(context.Background())
	assert.Error(t, err)
}

func TestAPI_GetUser(t *testing.T) {
	responseJSON := `{
		"user_id": "testuser",
		"display_name": "Test User",
		"email": "test@example.com",
		"type": "rgw",
		"tenant": "mytenant",
		"keys": [{"user": "testuser", "access_key": "key1", "secret_key": "secret1"}]
	}`

	mock := &mockHTTPClient{
		response: newMockResponse(200, responseJSON),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	user, err := api.GetUser(context.Background(), User{ID: "testuser"})
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.ID)
	assert.Equal(t, "Test User", user.DisplayName)
	assert.Equal(t, "mytenant", user.Tenant)
}

func TestAPI_GetUser_MissingID(t *testing.T) {
	mock := &mockHTTPClient{}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	_, err = api.GetUser(context.Background(), User{})
	assert.Error(t, err)
	assert.Equal(t, errMissingUserID, err)
}

func TestAPI_GetKVUser(t *testing.T) {
	responseJSON := `{
		"user_id": "testuser",
		"display_name": "Test User",
		"tenant": "t1"
	}`

	mock := &mockHTTPClient{
		response: newMockResponse(200, responseJSON),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	kvUser, err := api.GetKVUser(context.Background(), User{ID: "testuser"})
	assert.NoError(t, err)
	assert.Equal(t, "testuser", kvUser.ID)
	assert.Equal(t, "t1", kvUser.Tenant)
}

func TestAPI_ListBuckets(t *testing.T) {
	mock := &mockHTTPClient{
		response: newMockResponse(200, `["bucket1","bucket2","bucket3"]`),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	buckets, err := api.ListBuckets(context.Background())
	assert.NoError(t, err)
	assert.Len(t, buckets, 3)
	assert.Equal(t, "bucket1", buckets[0])
}

func TestAPI_GetBucketInfo(t *testing.T) {
	responseJSON := `{
		"bucket": "test-bucket",
		"owner": "admin",
		"tenant": "default",
		"id": "bucket-123"
	}`

	mock := &mockHTTPClient{
		response: newMockResponse(200, responseJSON),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	bucket, err := api.GetBucketInfo(context.Background(), Bucket{Bucket: "test-bucket"})
	assert.NoError(t, err)
	assert.Equal(t, "test-bucket", bucket.Bucket)
	assert.Equal(t, "admin", bucket.Owner)
}

func TestAPI_GetUsage(t *testing.T) {
	responseJSON := `{
		"entries": [
			{
				"user": "admin",
				"buckets": [
					{
						"bucket": "b1",
						"time": "2024-01-15 12:00:00.000000Z",
						"epoch": 1705320000,
						"owner": "admin",
						"categories": [
							{"category": "get_obj", "bytes_sent": 1024, "ops": 10, "successful_ops": 10}
						]
					}
				]
			}
		],
		"summary": []
	}`

	mock := &mockHTTPClient{
		response: newMockResponse(200, responseJSON),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	showEntries := true
	usage, err := api.GetUsage(context.Background(), Usage{
		UserID:      "admin",
		ShowEntries: &showEntries,
	})
	assert.NoError(t, err)
	assert.Len(t, usage.Entries, 1)
	assert.Equal(t, "admin", usage.Entries[0].User)
}

func TestAPI_ListBuckets_NotFound(t *testing.T) {
	mock := &mockHTTPClient{
		response: newMockResponse(404, `{"Code":"NoSuchBucket","RequestId":"r","HostId":"h"}`),
	}

	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)

	_, err = api.ListBuckets(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoSuchBucket))
}

func TestNew_WithCustomHTTPClient(t *testing.T) {
	mock := &mockHTTPClient{}
	api, err := New("http://localhost:8080", "access", "secret", mock)
	assert.NoError(t, err)
	assert.Equal(t, mock, api.HTTPClient)
}

func TestExplicitPlacement_JSON(t *testing.T) {
	jsonData := `{
		"data_pool": "default.rgw.buckets.data",
		"data_extra_pool": "default.rgw.buckets.non-ec",
		"index_pool": "default.rgw.buckets.index"
	}`

	var ep ExplicitPlacement
	err := json.Unmarshal([]byte(jsonData), &ep)
	assert.NoError(t, err)
	assert.Equal(t, "default.rgw.buckets.data", ep.DataPool)
	assert.Equal(t, "default.rgw.buckets.index", ep.IndexPool)
}

func TestSubuserAccessConstants(t *testing.T) {
	assert.Equal(t, SubuserAccess(""), SubuserAccessNone)
	assert.Equal(t, SubuserAccess("read"), SubuserAccessRead)
	assert.Equal(t, SubuserAccess("write"), SubuserAccessWrite)
	assert.Equal(t, SubuserAccess("readwrite"), SubuserAccessReadWrite)
	assert.Equal(t, SubuserAccess("full"), SubuserAccessFull)
	assert.Equal(t, SubuserAccess("full-control"), SubuserAccessReplyFull)
}
