// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package resourceusage

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceUsage_StructFields(t *testing.T) {
	usage := ResourceUsage{
		CPUUsage:    45.5,
		MemoryUsage: 72.3,
		DiskIO:      1024000,
		NetworkIO:   2048000,
		NodeName:    "node1",
		InstanceID:  "instance1",
	}

	assert.Equal(t, 45.5, usage.CPUUsage)
	assert.Equal(t, 72.3, usage.MemoryUsage)
	assert.Equal(t, uint64(1024000), usage.DiskIO)
	assert.Equal(t, uint64(2048000), usage.NetworkIO)
	assert.Equal(t, "node1", usage.NodeName)
	assert.Equal(t, "instance1", usage.InstanceID)
}

func TestResourceUsage_JSONSerialization(t *testing.T) {
	usage := ResourceUsage{
		CPUUsage:    45.5,
		MemoryUsage: 72.3,
		DiskIO:      1024000,
		NetworkIO:   2048000,
		NodeName:    "node1",
		InstanceID:  "instance1",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify JSON contains expected fields
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Equal(t, 45.5, result["cpu_usage"])
	assert.Equal(t, 72.3, result["memory_usage"])
	assert.Equal(t, float64(1024000), result["disk_io"])
	assert.Equal(t, float64(2048000), result["network_io"])
	assert.Equal(t, "node1", result["node_name"])
	assert.Equal(t, "instance1", result["instance_id"])
}

func TestResourceUsage_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"cpu_usage": 45.5,
		"memory_usage": 72.3,
		"disk_io": 1024000,
		"network_io": 2048000,
		"node_name": "node1",
		"instance_id": "instance1"
	}`

	var usage ResourceUsage
	err := json.Unmarshal([]byte(jsonData), &usage)
	assert.NoError(t, err)

	assert.Equal(t, 45.5, usage.CPUUsage)
	assert.Equal(t, 72.3, usage.MemoryUsage)
	assert.Equal(t, uint64(1024000), usage.DiskIO)
	assert.Equal(t, uint64(2048000), usage.NetworkIO)
	assert.Equal(t, "node1", usage.NodeName)
	assert.Equal(t, "instance1", usage.InstanceID)
}

func TestResourceUsage_ZeroValues(t *testing.T) {
	usage := ResourceUsage{}

	assert.Equal(t, 0.0, usage.CPUUsage)
	assert.Equal(t, 0.0, usage.MemoryUsage)
	assert.Equal(t, uint64(0), usage.DiskIO)
	assert.Equal(t, uint64(0), usage.NetworkIO)
	assert.Empty(t, usage.NodeName)
	assert.Empty(t, usage.InstanceID)
}

func TestResourceUsage_HighValues(t *testing.T) {
	// Test with high CPU and memory usage
	usage := ResourceUsage{
		CPUUsage:    99.9,
		MemoryUsage: 100.0,
		DiskIO:      18446744073709551615, // Max uint64
		NetworkIO:   18446744073709551615,
	}

	assert.Equal(t, 99.9, usage.CPUUsage)
	assert.Equal(t, 100.0, usage.MemoryUsage)
	assert.Equal(t, uint64(18446744073709551615), usage.DiskIO)
	assert.Equal(t, uint64(18446744073709551615), usage.NetworkIO)
}

func TestResourceUsageConfig_StructFields(t *testing.T) {
	config := ResourceUsageConfig{
		NatsURL:        "nats://localhost:4222",
		NatsSubject:    "resource.usage",
		UseNats:        true,
		Prometheus:     true,
		PrometheusPort: 9090,
		Interval:       30,
		Disks:          []string{"sda", "sdb", "nvme0n1"},
		NodeName:       "node1",
		InstanceID:     "instance1",
	}

	assert.Equal(t, "nats://localhost:4222", config.NatsURL)
	assert.Equal(t, "resource.usage", config.NatsSubject)
	assert.True(t, config.UseNats)
	assert.True(t, config.Prometheus)
	assert.Equal(t, 9090, config.PrometheusPort)
	assert.Equal(t, 30, config.Interval)
	assert.Equal(t, []string{"sda", "sdb", "nvme0n1"}, config.Disks)
	assert.Equal(t, "node1", config.NodeName)
	assert.Equal(t, "instance1", config.InstanceID)
}

func TestResourceUsageConfig_DefaultValues(t *testing.T) {
	config := ResourceUsageConfig{}

	assert.Empty(t, config.NatsURL)
	assert.Empty(t, config.NatsSubject)
	assert.False(t, config.UseNats)
	assert.False(t, config.Prometheus)
	assert.Equal(t, 0, config.PrometheusPort)
	assert.Equal(t, 0, config.Interval)
	assert.Nil(t, config.Disks)
	assert.Empty(t, config.NodeName)
	assert.Empty(t, config.InstanceID)
}

func TestResourceUsageConfig_MinimalConfig(t *testing.T) {
	// Test a minimal configuration for standalone mode
	config := ResourceUsageConfig{
		Interval: 60,
		Disks:    []string{"sda"},
	}

	assert.False(t, config.UseNats)
	assert.False(t, config.Prometheus)
	assert.Equal(t, 60, config.Interval)
	assert.Len(t, config.Disks, 1)
}

func TestResourceUsageConfig_EmptyDisks(t *testing.T) {
	config := ResourceUsageConfig{
		Interval: 30,
		Disks:    []string{},
	}

	assert.Empty(t, config.Disks)
}

func TestResourceUsageConfig_MultipleDisks(t *testing.T) {
	config := ResourceUsageConfig{
		Disks: []string{"sda", "sdb", "sdc", "nvme0n1", "nvme1n1"},
	}

	assert.Len(t, config.Disks, 5)
	assert.Contains(t, config.Disks, "sda")
	assert.Contains(t, config.Disks, "nvme0n1")
}

func TestResourceUsageConfig_NatsOnlyConfig(t *testing.T) {
	config := ResourceUsageConfig{
		NatsURL:     "nats://localhost:4222",
		NatsSubject: "resource.usage",
		UseNats:     true,
		Interval:    30,
		Disks:       []string{"sda"},
	}

	assert.True(t, config.UseNats)
	assert.False(t, config.Prometheus)
	assert.Equal(t, "nats://localhost:4222", config.NatsURL)
}

func TestResourceUsageConfig_PrometheusOnlyConfig(t *testing.T) {
	config := ResourceUsageConfig{
		Prometheus:     true,
		PrometheusPort: 8080,
		Interval:       30,
		Disks:          []string{"sda"},
	}

	assert.False(t, config.UseNats)
	assert.True(t, config.Prometheus)
	assert.Equal(t, 8080, config.PrometheusPort)
}
