// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"os"
	"testing"

	"github.com/cobaltcore-dev/prysm/pkg/producers/diskhealthmetrics"
	"github.com/cobaltcore-dev/prysm/pkg/producers/radosgwusage"
	"github.com/stretchr/testify/assert"
)

func TestMergeDiskHealthMetricsConfigWithEnv(t *testing.T) {
	// Clean up env vars after test
	envVars := []string{
		"NATS_URL", "NATS_SUBJECT", "PROMETHEUS_PORT", "ALL_ATTR",
		"DISKS", "NODE_NAME", "INSTANCE_ID", "INCLUDE_ZERO_VALUES",
		"INTERVAL", "GROWN_DEFECTS_THRESHOLD", "PENDING_SECTORS_THRESHOLD",
		"REALLOCATED_SECTORS_THRESHOLD", "LIFETIME_USED_THRESHOLD",
		"CEPH_OSD_BASE_PATH", "TEST_MODE", "TEST_DATA_PATH",
		"TEST_SCENARIO", "TEST_DEVICES",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	t.Run("uses defaults when no env vars set", func(t *testing.T) {
		cfg := diskhealthmetrics.DiskHealthMetricsConfig{
			NatsURL:     "default-url",
			NatsSubject: "default-subject",
			Interval:    60,
		}

		result := mergeDiskHealthMetricsConfigWithEnv(cfg)
		assert.Equal(t, "default-url", result.NatsURL)
		assert.Equal(t, "default-subject", result.NatsSubject)
		assert.Equal(t, 60, result.Interval)
	})

	t.Run("overrides with env vars", func(t *testing.T) {
		os.Setenv("NATS_URL", "nats://env-host:4222")
		os.Setenv("NODE_NAME", "env-node")
		os.Setenv("INTERVAL", "30")
		os.Setenv("DISKS", "sda,sdb,nvme0n1")
		os.Setenv("GROWN_DEFECTS_THRESHOLD", "50")
		os.Setenv("TEST_MODE", "true")
		os.Setenv("TEST_DEVICES", "test_sda,test_sdb")
		defer func() {
			for _, v := range envVars {
				os.Unsetenv(v)
			}
		}()

		cfg := diskhealthmetrics.DiskHealthMetricsConfig{}
		result := mergeDiskHealthMetricsConfigWithEnv(cfg)

		assert.Equal(t, "nats://env-host:4222", result.NatsURL)
		assert.Equal(t, "env-node", result.NodeName)
		assert.Equal(t, 30, result.Interval)
		assert.Equal(t, []string{"sda", "sdb", "nvme0n1"}, result.Disks)
		assert.Equal(t, int64(50), result.GrownDefectsThreshold)
		assert.True(t, result.TestMode)
		assert.Equal(t, []string{"test_sda", "test_sdb"}, result.TestDevices)
	})
}

func TestMergeRadosGWUsageConfigWithEnv(t *testing.T) {
	envVars := []string{
		"ADMIN_URL", "ACCESS_KEY", "SECRET_KEY", "NODE_NAME",
		"INSTANCE_ID", "PROMETHEUS_ENABLED", "PROMETHEUS_PORT",
		"COOLDOWN_INTERVAL", "RGW_CLUSTER_ID",
		"SYNC_CONTROL_NATS", "SYNC_EXTERNAL_NATS",
		"SYNC_CONTROL_URL", "SYNC_CONTROL_BUCKET_PREFIX",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	t.Run("uses defaults when no env vars set", func(t *testing.T) {
		cfg := radosgwusage.RadosGWUsageConfig{
			AdminURL:         "http://default:8080",
			PrometheusPort:   9090,
			CooldownInterval: 120,
		}

		result := mergeRadosGWUsageConfigWithEnv(cfg)
		assert.Equal(t, "http://default:8080", result.AdminURL)
		assert.Equal(t, 9090, result.PrometheusPort)
		assert.Equal(t, 120, result.CooldownInterval)
	})

	t.Run("overrides with env vars", func(t *testing.T) {
		os.Setenv("ADMIN_URL", "http://env-rgw:8080")
		os.Setenv("ACCESS_KEY", "env-access")
		os.Setenv("SECRET_KEY", "env-secret")
		os.Setenv("PROMETHEUS_PORT", "8888")
		os.Setenv("COOLDOWN_INTERVAL", "60")
		os.Setenv("RGW_CLUSTER_ID", "cluster-1")
		os.Setenv("SYNC_CONTROL_NATS", "true")
		defer func() {
			for _, v := range envVars {
				os.Unsetenv(v)
			}
		}()

		cfg := radosgwusage.RadosGWUsageConfig{}
		result := mergeRadosGWUsageConfigWithEnv(cfg)

		assert.Equal(t, "http://env-rgw:8080", result.AdminURL)
		assert.Equal(t, "env-access", result.AccessKey)
		assert.Equal(t, "env-secret", result.SecretKey)
		assert.Equal(t, 8888, result.PrometheusPort)
		assert.Equal(t, 60, result.CooldownInterval)
		assert.Equal(t, "cluster-1", result.ClusterID)
		assert.True(t, result.SyncControlNats)
	})
}
