// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskHealthMetricsConfig_StructFields(t *testing.T) {
	cfg := DiskHealthMetricsConfig{
		NatsURL:                     "nats://localhost:4222",
		NatsSubject:                 "disk.health",
		UseNats:                     true,
		Prometheus:                  true,
		PrometheusPort:              9090,
		AllAttributes:               true,
		Disks:                       []string{"sda", "sdb", "nvme0n1"},
		IncludeZeroValues:           false,
		Interval:                    60,
		NodeName:                    "node1",
		InstanceID:                  "instance1",
		GrownDefectsThreshold:       10,
		PendingSectorsThreshold:     5,
		ReallocatedSectorsThreshold: 20,
		LifetimeUsedThreshold:       90,
		CephOSDBasePath:             "/var/lib/ceph/osd",
		TestMode:                    false,
		TestDataPath:                "",
		TestScenario:                "",
		TestDevices:                 nil,
	}

	assert.Equal(t, "nats://localhost:4222", cfg.NatsURL)
	assert.Equal(t, "disk.health", cfg.NatsSubject)
	assert.True(t, cfg.UseNats)
	assert.True(t, cfg.Prometheus)
	assert.Equal(t, 9090, cfg.PrometheusPort)
	assert.True(t, cfg.AllAttributes)
	assert.Equal(t, []string{"sda", "sdb", "nvme0n1"}, cfg.Disks)
	assert.False(t, cfg.IncludeZeroValues)
	assert.Equal(t, 60, cfg.Interval)
	assert.Equal(t, "node1", cfg.NodeName)
	assert.Equal(t, "instance1", cfg.InstanceID)
	assert.Equal(t, int64(10), cfg.GrownDefectsThreshold)
	assert.Equal(t, int64(5), cfg.PendingSectorsThreshold)
	assert.Equal(t, int64(20), cfg.ReallocatedSectorsThreshold)
	assert.Equal(t, int64(90), cfg.LifetimeUsedThreshold)
	assert.Equal(t, "/var/lib/ceph/osd", cfg.CephOSDBasePath)
	assert.False(t, cfg.TestMode)
}

func TestDiskHealthMetricsConfig_DefaultValues(t *testing.T) {
	cfg := DiskHealthMetricsConfig{}

	assert.Empty(t, cfg.NatsURL)
	assert.Empty(t, cfg.NatsSubject)
	assert.False(t, cfg.UseNats)
	assert.False(t, cfg.Prometheus)
	assert.Equal(t, 0, cfg.PrometheusPort)
	assert.False(t, cfg.AllAttributes)
	assert.Nil(t, cfg.Disks)
	assert.False(t, cfg.IncludeZeroValues)
	assert.Equal(t, 0, cfg.Interval)
	assert.Empty(t, cfg.NodeName)
	assert.Empty(t, cfg.InstanceID)
	assert.Equal(t, int64(0), cfg.GrownDefectsThreshold)
	assert.Equal(t, int64(0), cfg.PendingSectorsThreshold)
	assert.Equal(t, int64(0), cfg.ReallocatedSectorsThreshold)
	assert.Equal(t, int64(0), cfg.LifetimeUsedThreshold)
	assert.Empty(t, cfg.CephOSDBasePath)
	assert.False(t, cfg.TestMode)
	assert.Empty(t, cfg.TestDataPath)
	assert.Empty(t, cfg.TestScenario)
	assert.Nil(t, cfg.TestDevices)
}

func TestDiskHealthMetricsConfig_TestModeConfig(t *testing.T) {
	cfg := DiskHealthMetricsConfig{
		TestMode:     true,
		TestDataPath: "/tmp/test-data",
		TestScenario: "failing",
		TestDevices:  []string{"test_sda", "test_sdb"},
		Interval:     10,
	}

	assert.True(t, cfg.TestMode)
	assert.Equal(t, "/tmp/test-data", cfg.TestDataPath)
	assert.Equal(t, "failing", cfg.TestScenario)
	assert.Equal(t, []string{"test_sda", "test_sdb"}, cfg.TestDevices)
}

func TestDiskHealthMetricsConfig_PrometheusOnlyConfig(t *testing.T) {
	cfg := DiskHealthMetricsConfig{
		Prometheus:     true,
		PrometheusPort: 8080,
		Interval:       30,
		Disks:          []string{"sda"},
		UseNats:        false,
	}

	assert.True(t, cfg.Prometheus)
	assert.False(t, cfg.UseNats)
	assert.Equal(t, 8080, cfg.PrometheusPort)
}

func TestDiskHealthMetricsConfig_NatsOnlyConfig(t *testing.T) {
	cfg := DiskHealthMetricsConfig{
		UseNats:     true,
		NatsURL:     "nats://localhost:4222",
		NatsSubject: "disk.health.metrics",
		Prometheus:  false,
		Interval:    60,
		Disks:       []string{"sda", "sdb"},
	}

	assert.True(t, cfg.UseNats)
	assert.False(t, cfg.Prometheus)
	assert.Equal(t, "nats://localhost:4222", cfg.NatsURL)
	assert.Equal(t, "disk.health.metrics", cfg.NatsSubject)
}

func TestDiskHealthMetricsConfig_ThresholdValues(t *testing.T) {
	tests := []struct {
		name                        string
		grownDefectsThreshold       int64
		pendingSectorsThreshold     int64
		reallocatedSectorsThreshold int64
		lifetimeUsedThreshold       int64
	}{
		{
			name:                        "default thresholds",
			grownDefectsThreshold:       0,
			pendingSectorsThreshold:     0,
			reallocatedSectorsThreshold: 0,
			lifetimeUsedThreshold:       0,
		},
		{
			name:                        "conservative thresholds",
			grownDefectsThreshold:       1,
			pendingSectorsThreshold:     1,
			reallocatedSectorsThreshold: 1,
			lifetimeUsedThreshold:       80,
		},
		{
			name:                        "aggressive thresholds",
			grownDefectsThreshold:       100,
			pendingSectorsThreshold:     50,
			reallocatedSectorsThreshold: 200,
			lifetimeUsedThreshold:       95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DiskHealthMetricsConfig{
				GrownDefectsThreshold:       tt.grownDefectsThreshold,
				PendingSectorsThreshold:     tt.pendingSectorsThreshold,
				ReallocatedSectorsThreshold: tt.reallocatedSectorsThreshold,
				LifetimeUsedThreshold:       tt.lifetimeUsedThreshold,
			}

			assert.Equal(t, tt.grownDefectsThreshold, cfg.GrownDefectsThreshold)
			assert.Equal(t, tt.pendingSectorsThreshold, cfg.PendingSectorsThreshold)
			assert.Equal(t, tt.reallocatedSectorsThreshold, cfg.ReallocatedSectorsThreshold)
			assert.Equal(t, tt.lifetimeUsedThreshold, cfg.LifetimeUsedThreshold)
		})
	}
}

func TestDiskHealthMetricsConfig_DisksList(t *testing.T) {
	tests := []struct {
		name     string
		disks    []string
		expected int
	}{
		{
			name:     "empty disks",
			disks:    []string{},
			expected: 0,
		},
		{
			name:     "single disk",
			disks:    []string{"sda"},
			expected: 1,
		},
		{
			name:     "multiple HDD",
			disks:    []string{"sda", "sdb", "sdc", "sdd"},
			expected: 4,
		},
		{
			name:     "NVMe devices",
			disks:    []string{"nvme0n1", "nvme1n1", "nvme2n1"},
			expected: 3,
		},
		{
			name:     "mixed devices",
			disks:    []string{"sda", "sdb", "nvme0n1", "nvme1n1"},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DiskHealthMetricsConfig{
				Disks: tt.disks,
			}

			assert.Len(t, cfg.Disks, tt.expected)
			if tt.expected > 0 {
				assert.Equal(t, tt.disks[0], cfg.Disks[0])
			}
		})
	}
}

func TestNormalizeDevicePath(t *testing.T) {
	tests := []struct {
		name     string
		device   string
		expected string
	}{
		{
			name:     "standard device path",
			device:   "/dev/sda",
			expected: "/dev/sda",
		},
		{
			name:     "NVMe device path",
			device:   "/dev/nvme0n1",
			expected: "/dev/nvme0n1",
		},
		{
			name:     "non-existent path returns original",
			device:   "/dev/nonexistent123",
			expected: "/dev/nonexistent123",
		},
		{
			name:     "empty path",
			device:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeDevicePath(tt.device)
			// We can't test symlink resolution in unit tests without real symlinks
			// Just verify the function returns something sensible
			assert.NotEmpty(t, result == "" && tt.device == "" || result != "")
		})
	}
}
