// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package radosgwusage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRadosGWUsageConfig_StructFields(t *testing.T) {
	cfg := RadosGWUsageConfig{
		AdminURL:                "http://localhost:8080",
		AccessKey:               "admin",
		SecretKey:               "secret",
		Prometheus:              true,
		PrometheusPort:          9090,
		NodeName:                "node1",
		InstanceID:              "instance1",
		CooldownInterval:        60,
		ClusterID:               "cluster-123",
		SyncControlNats:         true,
		SyncExternalNats:        false,
		SyncControlURL:          "nats://localhost:4222",
		SyncControlBucketPrefix: "prysm-sync",
	}

	assert.Equal(t, "http://localhost:8080", cfg.AdminURL)
	assert.Equal(t, "admin", cfg.AccessKey)
	assert.Equal(t, "secret", cfg.SecretKey)
	assert.True(t, cfg.Prometheus)
	assert.Equal(t, 9090, cfg.PrometheusPort)
	assert.Equal(t, "node1", cfg.NodeName)
	assert.Equal(t, "instance1", cfg.InstanceID)
	assert.Equal(t, 60, cfg.CooldownInterval)
	assert.Equal(t, "cluster-123", cfg.ClusterID)
	assert.True(t, cfg.SyncControlNats)
	assert.False(t, cfg.SyncExternalNats)
	assert.Equal(t, "nats://localhost:4222", cfg.SyncControlURL)
	assert.Equal(t, "prysm-sync", cfg.SyncControlBucketPrefix)
}

func TestRadosGWUsageConfig_DefaultValues(t *testing.T) {
	cfg := RadosGWUsageConfig{}

	assert.Empty(t, cfg.AdminURL)
	assert.Empty(t, cfg.AccessKey)
	assert.Empty(t, cfg.SecretKey)
	assert.False(t, cfg.Prometheus)
	assert.Equal(t, 0, cfg.PrometheusPort)
	assert.Empty(t, cfg.NodeName)
	assert.Empty(t, cfg.InstanceID)
	assert.Equal(t, 0, cfg.CooldownInterval)
	assert.Empty(t, cfg.ClusterID)
	assert.False(t, cfg.SyncControlNats)
	assert.False(t, cfg.SyncExternalNats)
	assert.Empty(t, cfg.SyncControlURL)
	assert.Empty(t, cfg.SyncControlBucketPrefix)
}

func TestRadosGWUsageConfig_MinimalConfig(t *testing.T) {
	// Minimal configuration for basic usage
	cfg := RadosGWUsageConfig{
		AdminURL:  "http://rgw:8080",
		AccessKey: "access",
		SecretKey: "secret",
	}

	assert.NotEmpty(t, cfg.AdminURL)
	assert.NotEmpty(t, cfg.AccessKey)
	assert.NotEmpty(t, cfg.SecretKey)
	assert.False(t, cfg.Prometheus)
	assert.False(t, cfg.SyncControlNats)
}

func TestRadosGWUsageConfig_WithSyncControl(t *testing.T) {
	// Configuration with sync control enabled
	cfg := RadosGWUsageConfig{
		AdminURL:                "http://rgw:8080",
		AccessKey:               "access",
		SecretKey:               "secret",
		SyncControlNats:         true,
		SyncExternalNats:        true,
		SyncControlURL:          "nats://external-nats:4222",
		SyncControlBucketPrefix: "my-prefix",
	}

	assert.True(t, cfg.SyncControlNats)
	assert.True(t, cfg.SyncExternalNats)
	assert.Equal(t, "nats://external-nats:4222", cfg.SyncControlURL)
	assert.Equal(t, "my-prefix", cfg.SyncControlBucketPrefix)
}
