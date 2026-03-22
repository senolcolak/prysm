// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package radosgwusage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRadosGWUserMetrics(t *testing.T) {
	metrics := NewRadosGWUserMetrics()

	// Verify Meta is initialized with empty strings
	assert.Empty(t, metrics.Meta.ID)
	assert.Empty(t, metrics.Meta.DisplayName)
	assert.Empty(t, metrics.Meta.Email)
	assert.Empty(t, metrics.Meta.DefaultStorageClass)

	// Verify Quota is initialized
	assert.False(t, metrics.Quota.Enabled)
	assert.Nil(t, metrics.Quota.MaxSize)
	assert.Nil(t, metrics.Quota.MaxObjects)

	// Verify Totals are initialized to zero
	assert.Equal(t, 0, metrics.Totals.BucketsTotal)
	assert.Equal(t, uint64(0), metrics.Totals.ObjectsTotal)
	assert.Equal(t, uint64(0), metrics.Totals.DataSizeTotal)
	assert.Equal(t, uint64(0), metrics.Totals.OpsTotal)
	assert.Equal(t, uint64(0), metrics.Totals.ReadOpsTotal)
	assert.Equal(t, uint64(0), metrics.Totals.WriteOpsTotal)
	assert.Equal(t, uint64(0), metrics.Totals.BytesSentTotal)
	assert.Equal(t, uint64(0), metrics.Totals.BytesReceivedTotal)
	assert.Equal(t, uint64(0), metrics.Totals.SuccessOpsTotal)
	assert.Equal(t, 0.0, metrics.Totals.ErrorRateTotal)
	assert.Equal(t, 0.0, metrics.Totals.ThroughputBytesTotal)
	assert.Equal(t, uint64(0), metrics.Totals.TotalCapacity)

	// Verify Current metrics are initialized to zero
	assert.Equal(t, 0.0, metrics.Current.OpsPerSec)
	assert.Equal(t, 0.0, metrics.Current.ReadOpsPerSec)
	assert.Equal(t, 0.0, metrics.Current.WriteOpsPerSec)
	assert.Equal(t, 0.0, metrics.Current.DataBytesReceivedPerSec)
	assert.Equal(t, 0.0, metrics.Current.DataBytesSentPerSec)
	assert.Equal(t, 0.0, metrics.Current.ThroughputBytesPerSec)
	assert.Equal(t, 0.0, metrics.Current.TotalAPIUsagePerSec)

	// Verify maps are initialized (not nil)
	assert.NotNil(t, metrics.APIUsagePerUser)
	assert.NotNil(t, metrics.Current.APIUsagePerSec)
	assert.Empty(t, metrics.APIUsagePerUser)
	assert.Empty(t, metrics.Current.APIUsagePerSec)
}

func TestNewRadosGWBucketMetrics(t *testing.T) {
	metrics := NewRadosGWBucketMetrics()

	// Verify Meta is initialized with empty strings
	assert.Empty(t, metrics.Meta.Name)
	assert.Empty(t, metrics.Meta.Owner)
	assert.Empty(t, metrics.Meta.Zonegroup)
	assert.Nil(t, metrics.Meta.Shards)
	assert.Nil(t, metrics.Meta.CreatedAt)

	// Verify Quota is initialized
	assert.False(t, metrics.Quota.Enabled)
	assert.Nil(t, metrics.Quota.MaxSize)
	assert.Nil(t, metrics.Quota.MaxObjects)

	// Verify Totals are initialized to zero
	assert.Equal(t, uint64(0), metrics.Totals.DataSize)
	assert.Equal(t, uint64(0), metrics.Totals.UtilizedSize)
	assert.Equal(t, uint64(0), metrics.Totals.Objects)
	assert.Equal(t, uint64(0), metrics.Totals.ReadOps)
	assert.Equal(t, uint64(0), metrics.Totals.WriteOps)
	assert.Equal(t, uint64(0), metrics.Totals.BytesSent)
	assert.Equal(t, uint64(0), metrics.Totals.BytesReceived)
	assert.Equal(t, uint64(0), metrics.Totals.SuccessOps)
	assert.Equal(t, uint64(0), metrics.Totals.OpsTotal)
	assert.Equal(t, 0.0, metrics.Totals.ErrorRate)
	assert.Equal(t, uint64(0), metrics.Totals.Capacity)

	// Verify Current metrics are initialized to zero
	assert.Equal(t, 0.0, metrics.Current.OpsPerSec)
	assert.Equal(t, 0.0, metrics.Current.ReadOpsPerSec)
	assert.Equal(t, 0.0, metrics.Current.WriteOpsPerSec)
	assert.Equal(t, 0.0, metrics.Current.BytesSentPerSec)
	assert.Equal(t, 0.0, metrics.Current.BytesReceivedPerSec)
	assert.Equal(t, 0.0, metrics.Current.ThroughputBytesPerSec)
	assert.Equal(t, 0.0, metrics.Current.TotalAPIUsagePerSec)

	// Verify maps are initialized
	assert.NotNil(t, metrics.Current.APIUsage)
	assert.Empty(t, metrics.Current.APIUsage)
}

func TestRadosGWUserMetrics_PopulateFields(t *testing.T) {
	metrics := NewRadosGWUserMetrics()

	// Populate metadata
	metrics.Meta.ID = "user123"
	metrics.Meta.DisplayName = "Test User"
	metrics.Meta.Email = "test@example.com"
	metrics.Meta.DefaultStorageClass = "STANDARD"

	// Populate quota
	maxSize := uint64(1073741824) // 1GB
	maxObjects := uint64(10000)
	metrics.Quota.Enabled = true
	metrics.Quota.MaxSize = &maxSize
	metrics.Quota.MaxObjects = &maxObjects

	// Populate totals
	metrics.Totals.BucketsTotal = 5
	metrics.Totals.ObjectsTotal = 1000
	metrics.Totals.DataSizeTotal = 536870912 // 512MB
	metrics.Totals.OpsTotal = 5000
	metrics.Totals.ReadOpsTotal = 3000
	metrics.Totals.WriteOpsTotal = 2000

	// Populate current
	metrics.Current.OpsPerSec = 10.5
	metrics.Current.ReadOpsPerSec = 6.5
	metrics.Current.WriteOpsPerSec = 4.0

	// Populate API usage
	metrics.APIUsagePerUser["get_obj"] = 100
	metrics.APIUsagePerUser["put_obj"] = 50

	// Verify
	assert.Equal(t, "user123", metrics.Meta.ID)
	assert.Equal(t, "Test User", metrics.Meta.DisplayName)
	assert.True(t, metrics.Quota.Enabled)
	assert.Equal(t, uint64(1073741824), *metrics.Quota.MaxSize)
	assert.Equal(t, 5, metrics.Totals.BucketsTotal)
	assert.Equal(t, 10.5, metrics.Current.OpsPerSec)
	assert.Equal(t, uint64(100), metrics.APIUsagePerUser["get_obj"])
}

func TestRadosGWBucketMetrics_PopulateFields(t *testing.T) {
	metrics := NewRadosGWBucketMetrics()

	// Populate metadata
	metrics.Meta.Name = "my-bucket"
	metrics.Meta.Owner = "user123"
	metrics.Meta.Zonegroup = "default"
	shards := uint64(8)
	metrics.Meta.Shards = &shards
	now := time.Now()
	metrics.Meta.CreatedAt = &now

	// Populate quota
	maxSize := uint64(10737418240) // 10GB
	metrics.Quota.Enabled = true
	metrics.Quota.MaxSize = &maxSize

	// Populate totals
	metrics.Totals.DataSize = 1073741824 // 1GB
	metrics.Totals.Objects = 500
	metrics.Totals.ReadOps = 1000
	metrics.Totals.WriteOps = 200
	metrics.Totals.OpsTotal = 1200
	metrics.Totals.ErrorRate = 2.5

	// Populate current
	metrics.Current.OpsPerSec = 5.0
	metrics.Current.ThroughputBytesPerSec = 1048576 // 1MB/s

	// Verify
	assert.Equal(t, "my-bucket", metrics.Meta.Name)
	assert.Equal(t, "user123", metrics.Meta.Owner)
	assert.Equal(t, uint64(8), *metrics.Meta.Shards)
	assert.NotNil(t, metrics.Meta.CreatedAt)
	assert.True(t, metrics.Quota.Enabled)
	assert.Equal(t, uint64(1073741824), metrics.Totals.DataSize)
	assert.Equal(t, 5.0, metrics.Current.OpsPerSec)
}

func TestUsageEntry_StructFields(t *testing.T) {
	entry := UsageEntry{
		ClusterID: "cluster-123",
	}

	assert.Equal(t, "cluster-123", entry.ClusterID)
}

func TestBucketUsage_StructFields(t *testing.T) {
	usage := BucketUsage{
		Bucket:               "test-bucket",
		Owner:                "test-owner",
		Zonegroup:            "default",
		NumShards:            8,
		TotalOps:             1000,
		TotalBytesSent:       1048576,
		TotalBytesReceived:   524288,
		TotalThroughputBytes: 1572864,
		TotalLatencySeconds:  0.5,
		TotalRequests:        1000,
		CurrentOps:           10,
		TotalReadOps:         600,
		TotalWriteOps:        400,
		TotalSuccessOps:      980,
		ErrorRate:            2.0,
	}

	assert.Equal(t, "test-bucket", usage.Bucket)
	assert.Equal(t, "test-owner", usage.Owner)
	assert.Equal(t, uint64(8), usage.NumShards)
	assert.Equal(t, uint64(1000), usage.TotalOps)
	assert.Equal(t, uint64(1048576), usage.TotalBytesSent)
	assert.Equal(t, 0.5, usage.TotalLatencySeconds)
	assert.Equal(t, 2.0, usage.ErrorRate)
}

func TestCategoryUsage_StructFields(t *testing.T) {
	category := CategoryUsage{
		Category:      "get_obj",
		BytesSent:     1048576,
		BytesReceived: 0,
		Ops:           100,
		SuccessfulOps: 98,
	}

	assert.Equal(t, "get_obj", category.Category)
	assert.Equal(t, uint64(1048576), category.BytesSent)
	assert.Equal(t, uint64(0), category.BytesReceived)
	assert.Equal(t, uint64(100), category.Ops)
	assert.Equal(t, uint64(98), category.SuccessfulOps)
}

func TestUsageMetrics_StructFields(t *testing.T) {
	metrics := UsageMetrics{
		Ops:           1000,
		SuccessfulOps: 980,
		BytesSent:     1048576,
		BytesReceived: 524288,
	}

	assert.Equal(t, uint64(1000), metrics.Ops)
	assert.Equal(t, uint64(980), metrics.SuccessfulOps)
	assert.Equal(t, uint64(1048576), metrics.BytesSent)
	assert.Equal(t, uint64(524288), metrics.BytesReceived)
}

func TestRadosGWClusterMetrics_StructFields(t *testing.T) {
	metrics := RadosGWClusterMetrics{
		OpsTotal:              10000,
		BytesSent:             10485760,
		BytesReceived:         5242880,
		ThroughputBytes:       15728640,
		ReadOpsPerSec:         100.5,
		WriteOpsPerSec:        50.3,
		BytesSentPerSec:       1048576,
		BytesReceivedPerSec:   524288,
		ThroughputBytesPerSec: 1572864,
		ErrorRate:             1.5,
		CurrentOpsPerSec:      150.8,
		CapacityUsageBytes:    107374182400,
	}

	assert.Equal(t, uint64(10000), metrics.OpsTotal)
	assert.Equal(t, 10485760.0, metrics.BytesSent)
	assert.Equal(t, 100.5, metrics.ReadOpsPerSec)
	assert.Equal(t, 1.5, metrics.ErrorRate)
	assert.Equal(t, uint64(107374182400), metrics.CapacityUsageBytes)
}

func TestUsageStats_JSONSerialization(t *testing.T) {
	size := uint64(1048576)
	numObjects := uint64(100)

	stats := UsageStats{}
	stats.RgwMain.Size = &size
	stats.RgwMain.NumObjects = &numObjects

	// Marshal to JSON
	jsonData, err := json.Marshal(stats)
	assert.NoError(t, err)

	// Unmarshal back
	var result UsageStats
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Equal(t, uint64(1048576), *result.RgwMain.Size)
	assert.Equal(t, uint64(100), *result.RgwMain.NumObjects)
}

func TestBucketUsage_JSONSerialization(t *testing.T) {
	usage := BucketUsage{
		Bucket:        "test-bucket",
		Owner:         "test-owner",
		TotalOps:      1000,
		TotalReadOps:  600,
		TotalWriteOps: 400,
		ErrorRate:     2.0,
		APIUsagePerBucket: map[string]int64{
			"get_obj": 500,
			"put_obj": 300,
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(usage)
	assert.NoError(t, err)

	// Unmarshal to map to verify field names
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Equal(t, "test-bucket", result["bucket"])
	assert.Equal(t, "test-owner", result["owner"])
	assert.Equal(t, float64(1000), result["total_ops"])
	assert.Equal(t, float64(600), result["read_ops"])
	assert.Equal(t, float64(400), result["write_ops"])
	assert.Equal(t, 2.0, result["error_rate"])
}

func TestCategoryUsage_JSONSerialization(t *testing.T) {
	categories := []CategoryUsage{
		{Category: "get_obj", Ops: 100, SuccessfulOps: 98, BytesSent: 1048576},
		{Category: "put_obj", Ops: 50, SuccessfulOps: 49, BytesReceived: 524288},
		{Category: "delete_obj", Ops: 20, SuccessfulOps: 20},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(categories)
	assert.NoError(t, err)

	// Unmarshal back
	var result []CategoryUsage
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Len(t, result, 3)
	assert.Equal(t, "get_obj", result[0].Category)
	assert.Equal(t, uint64(100), result[0].Ops)
	assert.Equal(t, "put_obj", result[1].Category)
	assert.Equal(t, uint64(524288), result[1].BytesReceived)
}

func TestRadosGWUserMetrics_APIUsageTracking(t *testing.T) {
	metrics := NewRadosGWUserMetrics()

	// Simulate API usage tracking
	metrics.APIUsagePerUser["get_obj"] = 1000
	metrics.APIUsagePerUser["put_obj"] = 500
	metrics.APIUsagePerUser["list_bucket"] = 200
	metrics.APIUsagePerUser["delete_obj"] = 50

	// Calculate totals
	var totalOps uint64
	for _, ops := range metrics.APIUsagePerUser {
		totalOps += ops
	}

	assert.Equal(t, uint64(1750), totalOps)
	assert.Equal(t, uint64(1000), metrics.APIUsagePerUser["get_obj"])
	assert.Equal(t, uint64(500), metrics.APIUsagePerUser["put_obj"])
}

func TestRadosGWBucketMetrics_APIUsageTracking(t *testing.T) {
	metrics := NewRadosGWBucketMetrics()

	// Initialize API usage map
	metrics.APIUsage = make(map[string]uint64)
	metrics.APIUsage["get_obj"] = 500
	metrics.APIUsage["put_obj"] = 250

	// Current API usage rate
	metrics.Current.APIUsage["get_obj"] = 10.5
	metrics.Current.APIUsage["put_obj"] = 5.2

	assert.Equal(t, uint64(500), metrics.APIUsage["get_obj"])
	assert.Equal(t, 10.5, metrics.Current.APIUsage["get_obj"])
}

func TestErrorRateCalculation(t *testing.T) {
	tests := []struct {
		name          string
		totalOps      uint64
		successOps    uint64
		expectedError float64
	}{
		{
			name:          "no errors",
			totalOps:      1000,
			successOps:    1000,
			expectedError: 0.0,
		},
		{
			name:          "2% error rate",
			totalOps:      1000,
			successOps:    980,
			expectedError: 2.0,
		},
		{
			name:          "50% error rate",
			totalOps:      100,
			successOps:    50,
			expectedError: 50.0,
		},
		{
			name:          "100% error rate",
			totalOps:      100,
			successOps:    0,
			expectedError: 100.0,
		},
		{
			name:          "zero operations",
			totalOps:      0,
			successOps:    0,
			expectedError: 0.0, // Avoid division by zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errorRate float64
			if tt.totalOps > 0 {
				errorRate = float64(tt.totalOps-tt.successOps) / float64(tt.totalOps) * 100
			}
			assert.InDelta(t, tt.expectedError, errorRate, 0.01)
		})
	}
}
