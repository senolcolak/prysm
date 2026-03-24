// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package radosgwusage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    *bool
		expected float64
	}{
		{
			name:     "nil returns 0",
			input:    nil,
			expected: 0.0,
		},
		{
			name:     "true returns 1",
			input:    boolPtr(true),
			expected: 1.0,
		},
		{
			name:     "false returns 0",
			input:    boolPtr(false),
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := boolToFloat64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestUserBucketMetrics_GetUserIdentification(t *testing.T) {
	tests := []struct {
		name     string
		metrics  UserBucketMetrics
		expected string
	}{
		{
			name:     "with tenant",
			metrics:  UserBucketMetrics{User: "alice", Tenant: "tenant1"},
			expected: "alice$tenant1",
		},
		{
			name:     "without tenant",
			metrics:  UserBucketMetrics{User: "alice", Tenant: ""},
			expected: "alice",
		},
		{
			name:     "empty user with tenant",
			metrics:  UserBucketMetrics{User: "", Tenant: "tenant1"},
			expected: "$tenant1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metrics.GetUserIdentification()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserLevelMetrics_GetUserIdentification(t *testing.T) {
	tests := []struct {
		name     string
		metrics  UserLevelMetrics
		expected string
	}{
		{
			name:     "with tenant",
			metrics:  UserLevelMetrics{User: "alice", Tenant: "tenant1"},
			expected: "alice$tenant1",
		},
		{
			name:     "without tenant",
			metrics:  UserLevelMetrics{User: "alice", Tenant: ""},
			expected: "alice",
		},
		{
			name:     "empty user with tenant",
			metrics:  UserLevelMetrics{User: "", Tenant: "tenant1"},
			expected: "$tenant1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metrics.GetUserIdentification()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserBucketMetrics_StructFields(t *testing.T) {
	numShards := uint64(16)
	quotaMaxSize := int64(1073741824)
	quotaMaxObjects := int64(10000)

	m := UserBucketMetrics{
		BucketID:        "bucket-1",
		User:            "alice",
		Tenant:          "tenant1",
		Zonegroup:       "default",
		ObjectCount:     100,
		BucketSize:      1024000,
		CreationTime:    "2024-01-15T12:00:00Z",
		NumShards:       &numShards,
		QuotaEnabled:    true,
		QuotaMaxSize:    &quotaMaxSize,
		QuotaMaxObjects: &quotaMaxObjects,
	}

	assert.Equal(t, "bucket-1", m.BucketID)
	assert.Equal(t, "alice", m.User)
	assert.Equal(t, uint64(100), m.ObjectCount)
	assert.Equal(t, uint64(1024000), m.BucketSize)
	assert.True(t, m.QuotaEnabled)
	assert.Equal(t, uint64(16), *m.NumShards)
}

func TestUserLevelMetrics_StructFields(t *testing.T) {
	quotaMaxSize := int64(5368709120)
	quotaMaxObjects := int64(50000)

	m := UserLevelMetrics{
		User:                "alice",
		Tenant:              "tenant1",
		DisplayName:         "Alice Smith",
		Email:               "alice@example.com",
		DefaultStorageClass: "STANDARD",
		Zonegroup:           "default",
		BucketsTotal:        5,
		ObjectsTotal:        1000,
		DataSizeTotal:       1024000000,
		UserQuotaEnabled:    true,
		UserQuotaMaxSize:    &quotaMaxSize,
		UserQuotaMaxObjects: &quotaMaxObjects,
	}

	assert.Equal(t, "alice", m.User)
	assert.Equal(t, "Alice Smith", m.DisplayName)
	assert.Equal(t, uint64(5), m.BucketsTotal)
	assert.Equal(t, uint64(1000), m.ObjectsTotal)
	assert.True(t, m.UserQuotaEnabled)
}
