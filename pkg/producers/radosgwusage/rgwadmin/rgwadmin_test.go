// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package rgwadmin

import (
	"encoding/json"
	"errors"
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
