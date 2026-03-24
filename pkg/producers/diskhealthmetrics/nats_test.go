// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeSSDWear(t *testing.T) {
	t.Run("media wearout indicator", func(t *testing.T) {
		attrs := map[string]SmartAttribute{
			"media_wearout_indicator": {RawValue: 90},
		}
		result := normalizeSSDWear(attrs)
		assert.NotNil(t, result)
		assert.Equal(t, int64(10), *result) // 100 - 90 = 10% used
	})

	t.Run("wear leveling count", func(t *testing.T) {
		attrs := map[string]SmartAttribute{
			"wear_leveling_count": {RawValue: 80},
		}
		result := normalizeSSDWear(attrs)
		assert.NotNil(t, result)
		assert.Equal(t, int64(20), *result) // 100 - 80 = 20% used
	})

	t.Run("no wear attribute found", func(t *testing.T) {
		attrs := map[string]SmartAttribute{
			"power_on_hours":     {RawValue: 5000},
			"temperature_celsius": {RawValue: 35},
		}
		result := normalizeSSDWear(attrs)
		assert.Nil(t, result)
	})

	t.Run("empty attributes", func(t *testing.T) {
		attrs := map[string]SmartAttribute{}
		result := normalizeSSDWear(attrs)
		assert.Nil(t, result)
	})

	t.Run("first matching attribute wins", func(t *testing.T) {
		attrs := map[string]SmartAttribute{
			"media_wearout_indicator": {RawValue: 90},
			"wear_leveling_count":     {RawValue: 80},
		}
		result := normalizeSSDWear(attrs)
		assert.NotNil(t, result)
		assert.Equal(t, int64(10), *result) // First match: media_wearout_indicator
	})
}

func TestGenerateMessage(t *testing.T) {
	tests := []struct {
		name     string
		details  map[string]string
		expected string
	}{
		{
			name:     "grown defects",
			details:  map[string]string{"GrownDefects": "10 (Warning)"},
			expected: "SMART data indicates potential drive issues (grown defects).",
		},
		{
			name:     "pending sectors",
			details:  map[string]string{"PendingSectors": "5 (Warning)"},
			expected: "SMART data indicates potential drive issues (pending sectors).",
		},
		{
			name:     "reallocated sectors",
			details:  map[string]string{"ReallocatedSectors": "20 (Warning)"},
			expected: "SMART data indicates potential drive issues (reallocated sectors).",
		},
		{
			name:     "SSD life used",
			details:  map[string]string{"SSDLifeUsed": "95% (Warning)"},
			expected: "SMART data indicates SSD nearing end of life.",
		},
		{
			name:     "no issues",
			details:  map[string]string{"Temperature": "35"},
			expected: "SMART data collected successfully.",
		},
		{
			name:     "empty details",
			details:  map[string]string{},
			expected: "SMART data collected successfully.",
		},
		{
			name:     "grown defects takes priority",
			details:  map[string]string{"GrownDefects": "10", "PendingSectors": "5"},
			expected: "SMART data indicates potential drive issues (grown defects).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateMessage(tt.details)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckAndSetThresholds(t *testing.T) {
	t.Run("no threshold exceeded", func(t *testing.T) {
		details := make(map[string]string)
		severity := "info"
		eventType := "health"

		normalizedData := NormalizedSmartData{
			Attributes: map[string]SmartAttribute{
				"grown_defects_count":    {RawValue: 0},
				"current_pending_sector": {RawValue: 0},
				"reallocated_sector_ct":  {RawValue: 0},
			},
		}
		config := &DiskHealthMetricsConfig{
			GrownDefectsThreshold:       10,
			PendingSectorsThreshold:     5,
			ReallocatedSectorsThreshold: 20,
			LifetimeUsedThreshold:       90,
		}

		checkAndSetThresholds(&details, normalizedData, config, &severity, &eventType)
		assert.Equal(t, "info", severity)
		assert.Equal(t, "health", eventType)
	})

	t.Run("grown defects threshold exceeded", func(t *testing.T) {
		details := make(map[string]string)
		severity := "info"
		eventType := "health"

		normalizedData := NormalizedSmartData{
			Attributes: map[string]SmartAttribute{
				"grown_defects_count":    {RawValue: 15},
				"current_pending_sector": {RawValue: 0},
				"reallocated_sector_ct":  {RawValue: 0},
			},
		}
		config := &DiskHealthMetricsConfig{
			GrownDefectsThreshold:       10,
			PendingSectorsThreshold:     5,
			ReallocatedSectorsThreshold: 20,
			LifetimeUsedThreshold:       90,
		}

		checkAndSetThresholds(&details, normalizedData, config, &severity, &eventType)
		assert.Equal(t, "warning", severity)
		assert.Equal(t, "health_alert", eventType)
		assert.Contains(t, details["GrownDefects"], "Warning")
	})

	t.Run("SSD lifetime threshold exceeded", func(t *testing.T) {
		details := make(map[string]string)
		severity := "info"
		eventType := "health"

		normalizedData := NormalizedSmartData{
			Attributes: map[string]SmartAttribute{
				"grown_defects_count":     {RawValue: 0},
				"current_pending_sector":  {RawValue: 0},
				"reallocated_sector_ct":   {RawValue: 0},
				"media_wearout_indicator": {RawValue: 5}, // 100-5=95% used
			},
		}
		config := &DiskHealthMetricsConfig{
			GrownDefectsThreshold:       10,
			PendingSectorsThreshold:     5,
			ReallocatedSectorsThreshold: 20,
			LifetimeUsedThreshold:       90,
		}

		checkAndSetThresholds(&details, normalizedData, config, &severity, &eventType)
		assert.Equal(t, "critical", severity)
		assert.Equal(t, "lifetime_alert", eventType)
	})
}

func TestConvertToNatsEvent(t *testing.T) {
	temp := int64(35)
	powerOnHours := int64(5000)

	normalizedData := NormalizedSmartData{
		NodeName:           "node1",
		InstanceID:         "instance1",
		Device:             "/dev/sda",
		TemperatureCelsius: &temp,
		PowerOnHours:       &powerOnHours,
		Attributes: map[string]SmartAttribute{
			"grown_defects_count":    {RawValue: 0},
			"current_pending_sector": {RawValue: 0},
			"reallocated_sector_ct":  {RawValue: 0},
		},
	}

	config := &DiskHealthMetricsConfig{
		GrownDefectsThreshold:       10,
		PendingSectorsThreshold:     5,
		ReallocatedSectorsThreshold: 20,
		LifetimeUsedThreshold:       90,
	}

	event := convertToNatsEvent(normalizedData, config)

	assert.Equal(t, "node1", event.NodeName)
	assert.Equal(t, "instance1", event.InstanceID)
	assert.Equal(t, "/dev/sda", event.Device)
	assert.Equal(t, "info", event.Severity)
	assert.Equal(t, "health", event.EventType)
	assert.Equal(t, "SMART data collected successfully.", event.Message)
	assert.Contains(t, event.Details["TemperatureCelsius"], "35")
	assert.Contains(t, event.Details["PowerOnHours"], "5000")
}

func TestConvertToNatsEvent_WithAlerts(t *testing.T) {
	normalizedData := NormalizedSmartData{
		NodeName:   "node1",
		InstanceID: "instance1",
		Device:     "/dev/sda",
		Attributes: map[string]SmartAttribute{
			"grown_defects_count":    {RawValue: 50},
			"current_pending_sector": {RawValue: 0},
			"reallocated_sector_ct":  {RawValue: 0},
		},
	}

	config := &DiskHealthMetricsConfig{
		GrownDefectsThreshold:       10,
		PendingSectorsThreshold:     5,
		ReallocatedSectorsThreshold: 20,
		LifetimeUsedThreshold:       90,
	}

	event := convertToNatsEvent(normalizedData, config)

	assert.Equal(t, "warning", event.Severity)
	assert.Equal(t, "health_alert", event.EventType)
	assert.Contains(t, event.Message, "grown defects")
}
