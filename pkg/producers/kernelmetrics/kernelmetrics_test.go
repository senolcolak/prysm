// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package kernelmetrics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKernelMetrics_StructFields(t *testing.T) {
	metrics := KernelMetrics{
		ContextSwitches: 123456,
		Entropy:         256,
		NetConnections:  100,
		NodeName:        "node1",
		InstanceID:      "instance1",
	}

	assert.Equal(t, uint64(123456), metrics.ContextSwitches)
	assert.Equal(t, uint64(256), metrics.Entropy)
	assert.Equal(t, uint64(100), metrics.NetConnections)
	assert.Equal(t, "node1", metrics.NodeName)
	assert.Equal(t, "instance1", metrics.InstanceID)
}

func TestKernelMetrics_JSONSerialization(t *testing.T) {
	metrics := KernelMetrics{
		ContextSwitches: 123456,
		Entropy:         256,
		NetConnections:  100,
		NodeName:        "node1",
		InstanceID:      "instance1",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(metrics)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify JSON contains expected fields
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Equal(t, float64(123456), result["context_switches"])
	assert.Equal(t, float64(256), result["entropy"])
	assert.Equal(t, float64(100), result["net_connections"])
	assert.Equal(t, "node1", result["node_name"])
	assert.Equal(t, "instance1", result["instance_id"])
}

func TestKernelMetrics_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"context_switches": 123456,
		"entropy": 256,
		"net_connections": 100,
		"node_name": "node1",
		"instance_id": "instance1"
	}`

	var metrics KernelMetrics
	err := json.Unmarshal([]byte(jsonData), &metrics)
	assert.NoError(t, err)

	assert.Equal(t, uint64(123456), metrics.ContextSwitches)
	assert.Equal(t, uint64(256), metrics.Entropy)
	assert.Equal(t, uint64(100), metrics.NetConnections)
	assert.Equal(t, "node1", metrics.NodeName)
	assert.Equal(t, "instance1", metrics.InstanceID)
}

func TestKernelMetrics_ZeroValues(t *testing.T) {
	metrics := KernelMetrics{}

	assert.Equal(t, uint64(0), metrics.ContextSwitches)
	assert.Equal(t, uint64(0), metrics.Entropy)
	assert.Equal(t, uint64(0), metrics.NetConnections)
	assert.Empty(t, metrics.NodeName)
	assert.Empty(t, metrics.InstanceID)
}

func TestKernelMetricsConfig_StructFields(t *testing.T) {
	config := KernelMetricsConfig{
		NatsURL:        "nats://localhost:4222",
		NatsSubject:    "kernel.metrics",
		UseNats:        true,
		NodeName:       "node1",
		InstanceID:     "instance1",
		Prometheus:     true,
		PrometheusPort: 9090,
		Interval:       30,
	}

	assert.Equal(t, "nats://localhost:4222", config.NatsURL)
	assert.Equal(t, "kernel.metrics", config.NatsSubject)
	assert.True(t, config.UseNats)
	assert.Equal(t, "node1", config.NodeName)
	assert.Equal(t, "instance1", config.InstanceID)
	assert.True(t, config.Prometheus)
	assert.Equal(t, 9090, config.PrometheusPort)
	assert.Equal(t, 30, config.Interval)
}

func TestKernelMetricsConfig_DefaultValues(t *testing.T) {
	config := KernelMetricsConfig{}

	assert.Empty(t, config.NatsURL)
	assert.Empty(t, config.NatsSubject)
	assert.False(t, config.UseNats)
	assert.Empty(t, config.NodeName)
	assert.Empty(t, config.InstanceID)
	assert.False(t, config.Prometheus)
	assert.Equal(t, 0, config.PrometheusPort)
	assert.Equal(t, 0, config.Interval)
}

func TestKernelMetricsConfig_MinimalConfig(t *testing.T) {
	// Test a minimal configuration for standalone mode (no NATS, no Prometheus)
	config := KernelMetricsConfig{
		Interval: 60,
	}

	assert.False(t, config.UseNats)
	assert.False(t, config.Prometheus)
	assert.Equal(t, 60, config.Interval)
}

func TestKernelMetricsConfig_NatsOnlyConfig(t *testing.T) {
	// Test configuration with NATS enabled but no Prometheus
	config := KernelMetricsConfig{
		NatsURL:     "nats://localhost:4222",
		NatsSubject: "kernel.metrics",
		UseNats:     true,
		Interval:    30,
	}

	assert.True(t, config.UseNats)
	assert.False(t, config.Prometheus)
	assert.Equal(t, "nats://localhost:4222", config.NatsURL)
}

func TestKernelMetricsConfig_PrometheusOnlyConfig(t *testing.T) {
	// Test configuration with Prometheus enabled but no NATS
	config := KernelMetricsConfig{
		Prometheus:     true,
		PrometheusPort: 8080,
		Interval:       30,
	}

	assert.False(t, config.UseNats)
	assert.True(t, config.Prometheus)
	assert.Equal(t, 8080, config.PrometheusPort)
}
