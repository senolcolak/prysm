// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package quotausageconsumer

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
	}

	assert.Equal(t, "user123", usage.UserID)
	assert.Equal(t, uint64(1073741824), usage.TotalQuota)
	assert.Equal(t, uint64(536870912), usage.UsedQuota)
	assert.Equal(t, uint64(536870912), usage.RemainingQuota)
	assert.Equal(t, "node1", usage.NodeName)
	assert.Equal(t, "instance1", usage.InstanceID)
}

func TestQuotaUsage_JSONSerialization(t *testing.T) {
	usage := QuotaUsage{
		UserID:         "user123",
		TotalQuota:     1073741824,
		UsedQuota:      536870912,
		RemainingQuota: 536870912,
		NodeName:       "node1",
		InstanceID:     "instance1",
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
}

func TestQuotaUsage_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"user_id": "user456",
		"total_quota": 2147483648,
		"used_quota": 1073741824,
		"remaining_quota": 1073741824,
		"node_name": "node2",
		"instance_id": "instance2"
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
}

func TestQuotaUsage_ZeroValues(t *testing.T) {
	usage := QuotaUsage{}

	assert.Empty(t, usage.UserID)
	assert.Equal(t, uint64(0), usage.TotalQuota)
	assert.Equal(t, uint64(0), usage.UsedQuota)
	assert.Equal(t, uint64(0), usage.RemainingQuota)
	assert.Empty(t, usage.NodeName)
	assert.Empty(t, usage.InstanceID)
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

func TestQuotaUsage_JSONArray(t *testing.T) {
	// Test deserializing an array of QuotaUsage (as received from NATS)
	jsonData := `[
		{
			"user_id": "user1",
			"total_quota": 1000000,
			"used_quota": 500000,
			"remaining_quota": 500000,
			"node_name": "node1",
			"instance_id": "inst1"
		},
		{
			"user_id": "user2",
			"total_quota": 2000000,
			"used_quota": 1800000,
			"remaining_quota": 200000,
			"node_name": "node1",
			"instance_id": "inst1"
		}
	]`

	var quotas []QuotaUsage
	err := json.Unmarshal([]byte(jsonData), &quotas)
	assert.NoError(t, err)

	assert.Len(t, quotas, 2)
	assert.Equal(t, "user1", quotas[0].UserID)
	assert.Equal(t, uint64(500000), quotas[0].UsedQuota)
	assert.Equal(t, "user2", quotas[1].UserID)
	assert.Equal(t, uint64(1800000), quotas[1].UsedQuota)
}

func TestQuotaUsageConsumerConfig_StructFields(t *testing.T) {
	cfg := QuotaUsageConsumerConfig{
		NatsURL:           "nats://localhost:4222",
		NatsSubject:       "quota.usage",
		Prometheus:        true,
		PrometheusPort:    9090,
		QuotaUsagePercent: 80.0,
		NodeName:          "node1",
		InstanceID:        "instance1",
	}

	assert.Equal(t, "nats://localhost:4222", cfg.NatsURL)
	assert.Equal(t, "quota.usage", cfg.NatsSubject)
	assert.True(t, cfg.Prometheus)
	assert.Equal(t, 9090, cfg.PrometheusPort)
	assert.Equal(t, 80.0, cfg.QuotaUsagePercent)
	assert.Equal(t, "node1", cfg.NodeName)
	assert.Equal(t, "instance1", cfg.InstanceID)
}

func TestQuotaUsageConsumerConfig_DefaultValues(t *testing.T) {
	cfg := QuotaUsageConsumerConfig{}

	assert.Empty(t, cfg.NatsURL)
	assert.Empty(t, cfg.NatsSubject)
	assert.False(t, cfg.Prometheus)
	assert.Equal(t, 0, cfg.PrometheusPort)
	assert.Equal(t, 0.0, cfg.QuotaUsagePercent)
	assert.Empty(t, cfg.NodeName)
	assert.Empty(t, cfg.InstanceID)
}

func TestQuotaUsageConsumerConfig_HighThreshold(t *testing.T) {
	cfg := QuotaUsageConsumerConfig{
		QuotaUsagePercent: 95.0, // Alert at 95% usage
	}

	assert.Equal(t, 95.0, cfg.QuotaUsagePercent)
}

func TestQuotaUsageConsumerConfig_LowThreshold(t *testing.T) {
	cfg := QuotaUsageConsumerConfig{
		QuotaUsagePercent: 50.0, // Alert at 50% usage
	}

	assert.Equal(t, 50.0, cfg.QuotaUsagePercent)
}

// Test usage percentage calculation logic
func TestCalculateUsagePercent(t *testing.T) {
	tests := []struct {
		name           string
		totalQuota     uint64
		usedQuota      uint64
		expectedPct    float64
		thresholdPct   float64
		shouldAlert    bool
	}{
		{
			name:           "50% usage below threshold",
			totalQuota:     1000,
			usedQuota:      500,
			expectedPct:    50.0,
			thresholdPct:   80.0,
			shouldAlert:    false,
		},
		{
			name:           "80% usage at threshold",
			totalQuota:     1000,
			usedQuota:      800,
			expectedPct:    80.0,
			thresholdPct:   80.0,
			shouldAlert:    true,
		},
		{
			name:           "90% usage above threshold",
			totalQuota:     1000,
			usedQuota:      900,
			expectedPct:    90.0,
			thresholdPct:   80.0,
			shouldAlert:    true,
		},
		{
			name:           "100% usage",
			totalQuota:     1000,
			usedQuota:      1000,
			expectedPct:    100.0,
			thresholdPct:   80.0,
			shouldAlert:    true,
		},
		{
			name:           "zero total quota",
			totalQuota:     0,
			usedQuota:      100,
			expectedPct:    0.0, // Should not calculate
			thresholdPct:   80.0,
			shouldAlert:    false,
		},
		{
			name:           "used exceeds total (capped)",
			totalQuota:     1000,
			usedQuota:      1500,
			expectedPct:    100.0, // Capped at total
			thresholdPct:   80.0,
			shouldAlert:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.totalQuota > 0 {
				usedQuota := tt.usedQuota
				if usedQuota > tt.totalQuota {
					usedQuota = tt.totalQuota
				}
				usagePercent := (float64(usedQuota) / float64(tt.totalQuota)) * 100
				assert.InDelta(t, tt.expectedPct, usagePercent, 0.01)
				assert.Equal(t, tt.shouldAlert, usagePercent >= tt.thresholdPct)
			}
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
	assert.Equal(t, uint64(0), usage.UsedQuota)         // Default
	assert.Equal(t, uint64(0), usage.RemainingQuota)    // Default
	assert.Empty(t, usage.NodeName)                     // Default
	assert.Empty(t, usage.InstanceID)                   // Default
}

func TestQuotaUsage_InvalidJSON(t *testing.T) {
	jsonData := `{invalid json}`

	var usage QuotaUsage
	err := json.Unmarshal([]byte(jsonData), &usage)
	assert.Error(t, err)
}

func TestQuotaUsage_EmptyArray(t *testing.T) {
	jsonData := `[]`

	var quotas []QuotaUsage
	err := json.Unmarshal([]byte(jsonData), &quotas)
	assert.NoError(t, err)
	assert.Empty(t, quotas)
}
