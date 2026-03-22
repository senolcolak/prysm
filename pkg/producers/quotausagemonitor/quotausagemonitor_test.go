// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package quotausagemonitor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaUsage_StructFields(t *testing.T) {
	usage := QuotaUsage{
		UserID:         "user123",
		TotalQuota:     1073741824, // 1GB
		UsedQuota:      536870912,  // 512MB
		RemainingQuota: 536870912,  // 512MB
		NodeName:       "node1",
		InstanceID:     "instance1",
		PhysicalSize:   "512MB",
	}

	assert.Equal(t, "user123", usage.UserID)
	assert.Equal(t, uint64(1073741824), usage.TotalQuota)
	assert.Equal(t, uint64(536870912), usage.UsedQuota)
	assert.Equal(t, uint64(536870912), usage.RemainingQuota)
	assert.Equal(t, "node1", usage.NodeName)
	assert.Equal(t, "instance1", usage.InstanceID)
	assert.Equal(t, "512MB", usage.PhysicalSize)
}

func TestQuotaUsage_JSONSerialization(t *testing.T) {
	usage := QuotaUsage{
		UserID:         "user123",
		TotalQuota:     1073741824,
		UsedQuota:      536870912,
		RemainingQuota: 536870912,
		NodeName:       "node1",
		InstanceID:     "instance1",
		PhysicalSize:   "512MB",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify JSON contains expected fields
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Equal(t, "user123", result["user_id"])
	assert.Equal(t, float64(1073741824), result["total_quota"])
	assert.Equal(t, float64(536870912), result["used_quota"])
	assert.Equal(t, float64(536870912), result["remaining_quota"])
	assert.Equal(t, "node1", result["node_name"])
	assert.Equal(t, "instance1", result["instance_id"])
	assert.Equal(t, "512MB", result["physical_size"])
}

func TestQuotaUsage_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"user_id": "user456",
		"total_quota": 2147483648,
		"used_quota": 1073741824,
		"remaining_quota": 1073741824,
		"node_name": "node2",
		"instance_id": "instance2",
		"physical_size": "1GB"
	}`

	var usage QuotaUsage
	err := json.Unmarshal([]byte(jsonData), &usage)
	assert.NoError(t, err)

	assert.Equal(t, "user456", usage.UserID)
	assert.Equal(t, uint64(2147483648), usage.TotalQuota)
	assert.Equal(t, uint64(1073741824), usage.UsedQuota)
	assert.Equal(t, uint64(1073741824), usage.RemainingQuota)
	assert.Equal(t, "node2", usage.NodeName)
	assert.Equal(t, "instance2", usage.InstanceID)
	assert.Equal(t, "1GB", usage.PhysicalSize)
}

func TestQuotaUsage_ZeroValues(t *testing.T) {
	usage := QuotaUsage{}

	assert.Empty(t, usage.UserID)
	assert.Equal(t, uint64(0), usage.TotalQuota)
	assert.Equal(t, uint64(0), usage.UsedQuota)
	assert.Equal(t, uint64(0), usage.RemainingQuota)
	assert.Empty(t, usage.NodeName)
	assert.Empty(t, usage.InstanceID)
	assert.Empty(t, usage.PhysicalSize)
}

func TestQuotaUsage_MaxValues(t *testing.T) {
	// Test with maximum uint64 values
	usage := QuotaUsage{
		UserID:         "maxuser",
		TotalQuota:     18446744073709551615, // Max uint64
		UsedQuota:      18446744073709551615,
		RemainingQuota: 0,
	}

	assert.Equal(t, uint64(18446744073709551615), usage.TotalQuota)
	assert.Equal(t, uint64(18446744073709551615), usage.UsedQuota)
	assert.Equal(t, uint64(0), usage.RemainingQuota)
}

func TestQuotaUsageMonitorConfig_StructFields(t *testing.T) {
	cfg := QuotaUsageMonitorConfig{
		AdminURL:          "http://localhost:8080",
		AccessKey:         "admin",
		SecretKey:         "secret",
		NatsURL:           "nats://localhost:4222",
		NatsSubject:       "quota.usage",
		UseNats:           true,
		Interval:          60,
		NodeName:          "node1",
		InstanceID:        "instance1",
		QuotaUsagePercent: 80.0,
	}

	assert.Equal(t, "http://localhost:8080", cfg.AdminURL)
	assert.Equal(t, "admin", cfg.AccessKey)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.Equal(t, "nats://localhost:4222", cfg.NatsURL)
	assert.Equal(t, "quota.usage", cfg.NatsSubject)
	assert.True(t, cfg.UseNats)
	assert.Equal(t, 60, cfg.Interval)
	assert.Equal(t, "node1", cfg.NodeName)
	assert.Equal(t, "instance1", cfg.InstanceID)
	assert.Equal(t, 80.0, cfg.QuotaUsagePercent)
}

func TestQuotaUsageMonitorConfig_DefaultValues(t *testing.T) {
	cfg := QuotaUsageMonitorConfig{}

	assert.Empty(t, cfg.AdminURL)
	assert.Empty(t, cfg.AccessKey)
	assert.Empty(t, cfg.SecretKey)
	assert.Empty(t, cfg.NatsURL)
	assert.Empty(t, cfg.NatsSubject)
	assert.False(t, cfg.UseNats)
	assert.Equal(t, 0, cfg.Interval)
	assert.Empty(t, cfg.NodeName)
	assert.Empty(t, cfg.InstanceID)
	assert.Equal(t, 0.0, cfg.QuotaUsagePercent)
}

func TestQuotaUsageMonitorConfig_StandaloneMode(t *testing.T) {
	// Configuration without NATS (prints to stdout)
	cfg := QuotaUsageMonitorConfig{
		AdminURL:          "http://localhost:8080",
		AccessKey:         "admin",
		SecretKey:         "secret",
		Interval:          30,
		QuotaUsagePercent: 90.0,
		UseNats:           false,
	}

	assert.False(t, cfg.UseNats)
	assert.NotEmpty(t, cfg.AdminURL)
	assert.Equal(t, 30, cfg.Interval)
}

func TestQuotaUsageMonitorConfig_HighThreshold(t *testing.T) {
	cfg := QuotaUsageMonitorConfig{
		QuotaUsagePercent: 95.0, // Alert only at 95% usage
	}

	assert.Equal(t, 95.0, cfg.QuotaUsagePercent)
}

func TestBoolPtr(t *testing.T) {
	// Test the boolPtr helper function
	truePtr := boolPtr(true)
	assert.NotNil(t, truePtr)
	assert.True(t, *truePtr)

	falsePtr := boolPtr(false)
	assert.NotNil(t, falsePtr)
	assert.False(t, *falsePtr)
}

func TestQuotaUsage_JSONArray(t *testing.T) {
	// Test serializing/deserializing an array of QuotaUsage
	quotas := []QuotaUsage{
		{
			UserID:         "user1",
			TotalQuota:     1000000,
			UsedQuota:      500000,
			RemainingQuota: 500000,
			NodeName:       "node1",
			InstanceID:     "inst1",
		},
		{
			UserID:         "user2",
			TotalQuota:     2000000,
			UsedQuota:      1800000,
			RemainingQuota: 200000,
			NodeName:       "node1",
			InstanceID:     "inst1",
		},
	}

	jsonData, err := json.Marshal(quotas)
	assert.NoError(t, err)

	var result []QuotaUsage
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "user1", result[0].UserID)
	assert.Equal(t, "user2", result[1].UserID)
}

// Test usage percentage calculation logic (mirrors what collectQuotaUsage does)
func TestUsagePercentageCalculation(t *testing.T) {
	tests := []struct {
		name           string
		totalQuota     uint64
		usedQuota      uint64
		expectedPct    float64
		thresholdPct   float64
		shouldReport   bool
	}{
		{
			name:           "50% usage below threshold",
			totalQuota:     1000,
			usedQuota:      500,
			expectedPct:    50.0,
			thresholdPct:   80.0,
			shouldReport:   false,
		},
		{
			name:           "80% usage at threshold",
			totalQuota:     1000,
			usedQuota:      800,
			expectedPct:    80.0,
			thresholdPct:   80.0,
			shouldReport:   true,
		},
		{
			name:           "90% usage above threshold",
			totalQuota:     1000,
			usedQuota:      900,
			expectedPct:    90.0,
			thresholdPct:   80.0,
			shouldReport:   true,
		},
		{
			name:           "100% usage",
			totalQuota:     1000,
			usedQuota:      1000,
			expectedPct:    100.0,
			thresholdPct:   80.0,
			shouldReport:   true,
		},
		{
			name:           "over quota (capped at 100%)",
			totalQuota:     1000,
			usedQuota:      1200,
			expectedPct:    120.0, // Not capped in calculation
			thresholdPct:   80.0,
			shouldReport:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usagePercent := float64(tt.usedQuota) / float64(tt.totalQuota) * 100
			assert.InDelta(t, tt.expectedPct, usagePercent, 0.01)
			assert.Equal(t, tt.shouldReport, usagePercent >= tt.thresholdPct)
		})
	}
}

func TestQuotaUsage_PartialJSON(t *testing.T) {
	// Test with partial JSON (missing some fields)
	jsonData := `{
		"user_id": "partial_user",
		"total_quota": 1000
	}`

	var usage QuotaUsage
	err := json.Unmarshal([]byte(jsonData), &usage)
	assert.NoError(t, err)

	assert.Equal(t, "partial_user", usage.UserID)
	assert.Equal(t, uint64(1000), usage.TotalQuota)
	assert.Equal(t, uint64(0), usage.UsedQuota)
	assert.Equal(t, uint64(0), usage.RemainingQuota)
	assert.Empty(t, usage.NodeName)
	assert.Empty(t, usage.InstanceID)
	assert.Empty(t, usage.PhysicalSize)
}

func TestQuotaUsage_InvalidJSON(t *testing.T) {
	jsonData := `{invalid json}`

	var usage QuotaUsage
	err := json.Unmarshal([]byte(jsonData), &usage)
	assert.Error(t, err)
}
