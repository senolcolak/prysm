// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"os"
	"testing"

	"github.com/cobaltcore-dev/prysm/pkg/consumer/quotausageconsumer"
	"github.com/cobaltcore-dev/prysm/pkg/producers/bucketnotify"
	"github.com/cobaltcore-dev/prysm/pkg/producers/diskhealthmetrics"
	"github.com/cobaltcore-dev/prysm/pkg/producers/kernelmetrics"
	"github.com/cobaltcore-dev/prysm/pkg/producers/opslog"
	"github.com/cobaltcore-dev/prysm/pkg/producers/quotausagemonitor"
	"github.com/cobaltcore-dev/prysm/pkg/producers/radosgwusage"
	"github.com/cobaltcore-dev/prysm/pkg/producers/resourceusage"
	"github.com/stretchr/testify/assert"
)

// helper to clear env vars
func clearEnvVars(vars []string) {
	for _, v := range vars {
		os.Unsetenv(v)
	}
}

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

func TestMergeOpsLogConfigWithEnv(t *testing.T) {
	envVars := []string{
		"LOG_FILE_PATH", "TRUNCATE_LOG_ON_START", "SOCKET_PATH",
		"NATS_URL", "NATS_SUBJECT", "NATS_METRICS_SUBJECT",
		"LOG_TO_STDOUT", "LOG_PRETTY_PRINT", "LOG_RETENTION_DAYS",
		"MAX_LOG_FILE_SIZE", "PROMETHEUS_PORT", "POD_NAME",
		"IGNORE_ANONYMOUS_REQUESTS", "PROMETHEUS_INTERVAL",
		"TRACK_EVERYTHING", "TRACK_REQUESTS_DETAILED",
	}
	clearEnvVars(envVars)

	t.Run("uses defaults", func(t *testing.T) {
		cfg := opslog.OpsLogConfig{
			LogFilePath:    "/var/log/default.log",
			PrometheusPort: 9090,
		}
		result := mergeOpsLogConfigWithEnv(cfg)
		assert.Equal(t, "/var/log/default.log", result.LogFilePath)
		assert.Equal(t, 9090, result.PrometheusPort)
	})

	t.Run("overrides with env", func(t *testing.T) {
		os.Setenv("LOG_FILE_PATH", "/tmp/ops.log")
		os.Setenv("PROMETHEUS_PORT", "7777")
		os.Setenv("NATS_URL", "nats://env:4222")
		os.Setenv("POD_NAME", "test-pod")
		os.Setenv("TRACK_EVERYTHING", "true")
		os.Setenv("PROMETHEUS_INTERVAL", "15")
		defer clearEnvVars(envVars)

		cfg := opslog.OpsLogConfig{}
		result := mergeOpsLogConfigWithEnv(cfg)

		assert.Equal(t, "/tmp/ops.log", result.LogFilePath)
		assert.Equal(t, 7777, result.PrometheusPort)
		assert.Equal(t, "nats://env:4222", result.NatsURL)
		assert.Equal(t, "test-pod", result.PodName)
		assert.True(t, result.MetricsConfig.TrackEverything)
		assert.Equal(t, 15, result.PrometheusIntervalSeconds)
	})
}

func TestMergeKernelMetricsConfigWithEnv(t *testing.T) {
	envVars := []string{"NATS_URL", "NATS_SUBJECT", "NODE_NAME", "INSTANCE_ID", "INTERVAL", "PROMETHEUS_PORT"}
	clearEnvVars(envVars)

	t.Run("uses defaults", func(t *testing.T) {
		cfg := kernelmetrics.KernelMetricsConfig{
			NatsURL:  "default-url",
			Interval: 30,
		}
		result := mergeKernelMetricsConfigWithEnv(cfg)
		assert.Equal(t, "default-url", result.NatsURL)
		assert.Equal(t, 30, result.Interval)
	})

	t.Run("overrides with env", func(t *testing.T) {
		os.Setenv("NATS_URL", "nats://env:4222")
		os.Setenv("NODE_NAME", "test-node")
		os.Setenv("INTERVAL", "15")
		os.Setenv("PROMETHEUS_PORT", "9999")
		defer clearEnvVars(envVars)

		cfg := kernelmetrics.KernelMetricsConfig{}
		result := mergeKernelMetricsConfigWithEnv(cfg)

		assert.Equal(t, "nats://env:4222", result.NatsURL)
		assert.Equal(t, "test-node", result.NodeName)
		assert.Equal(t, 15, result.Interval)
		assert.Equal(t, 9999, result.PrometheusPort)
	})
}

func TestMergeResourceUsageConfigWithEnv(t *testing.T) {
	envVars := []string{"NATS_URL", "NATS_SUBJECT", "PROMETHEUS_PORT", "DISKS", "NODE_NAME", "INSTANCE_ID", "INTERVAL"}
	clearEnvVars(envVars)

	t.Run("uses defaults", func(t *testing.T) {
		cfg := resourceusage.ResourceUsageConfig{
			NatsURL:  "default",
			Interval: 60,
		}
		result := mergeResourceUsageConfigWithEnv(cfg)
		assert.Equal(t, "default", result.NatsURL)
		assert.Equal(t, 60, result.Interval)
	})

	t.Run("overrides DISKS from env", func(t *testing.T) {
		os.Setenv("DISKS", "sda,sdb,nvme0n1")
		os.Setenv("NODE_NAME", "worker-1")
		defer clearEnvVars(envVars)

		cfg := resourceusage.ResourceUsageConfig{}
		result := mergeResourceUsageConfigWithEnv(cfg)

		assert.Equal(t, []string{"sda", "sdb", "nvme0n1"}, result.Disks)
		assert.Equal(t, "worker-1", result.NodeName)
	})
}

func TestMergeBucketNotifyConfigWithEnv(t *testing.T) {
	envVars := []string{"BUCKET_NOTIFY_ENDPOINT_PORT", "NATS_URL", "NATS_SUBJECT"}
	clearEnvVars(envVars)

	t.Run("uses defaults", func(t *testing.T) {
		cfg := bucketnotify.BucketNotifyConfig{
			EndpointPort: 8080,
			NatsURL:      "default-url",
		}
		result := mergeBucketNotifyConfigWithEnv(cfg)
		assert.Equal(t, 8080, result.EndpointPort)
		assert.Equal(t, "default-url", result.NatsURL)
	})

	t.Run("overrides with env", func(t *testing.T) {
		os.Setenv("BUCKET_NOTIFY_ENDPOINT_PORT", "9000")
		os.Setenv("NATS_URL", "nats://notify:4222")
		os.Setenv("NATS_SUBJECT", "bucket.events")
		defer clearEnvVars(envVars)

		cfg := bucketnotify.BucketNotifyConfig{}
		result := mergeBucketNotifyConfigWithEnv(cfg)

		assert.Equal(t, 9000, result.EndpointPort)
		assert.Equal(t, "nats://notify:4222", result.NatsURL)
		assert.Equal(t, "bucket.events", result.NatsSubject)
	})
}

func TestMergeQuotaUsageConsumerConfigWithEnv(t *testing.T) {
	envVars := []string{"NATS_URL", "NATS_SUBJECT", "PROMETHEUS_PORT", "QUOTA_USAGE_PERCENT", "NODE_NAME", "INSTANCE_ID"}
	clearEnvVars(envVars)

	t.Run("uses defaults", func(t *testing.T) {
		cfg := quotausageconsumer.QuotaUsageConsumerConfig{
			NatsURL:           "default",
			QuotaUsagePercent: 80.0,
		}
		result := mergeQuotaUsageConsumerConfigWithEnv(cfg)
		assert.Equal(t, "default", result.NatsURL)
		assert.Equal(t, 80.0, result.QuotaUsagePercent)
	})

	t.Run("overrides with env", func(t *testing.T) {
		os.Setenv("NATS_URL", "nats://consumer:4222")
		os.Setenv("QUOTA_USAGE_PERCENT", "95.5")
		os.Setenv("PROMETHEUS_PORT", "8181")
		defer clearEnvVars(envVars)

		cfg := quotausageconsumer.QuotaUsageConsumerConfig{}
		result := mergeQuotaUsageConsumerConfigWithEnv(cfg)

		assert.Equal(t, "nats://consumer:4222", result.NatsURL)
		assert.Equal(t, 95.5, result.QuotaUsagePercent)
		assert.Equal(t, 8181, result.PrometheusPort)
	})
}

func TestMergeQuotaUsageMonitorConfigWithEnv(t *testing.T) {
	envVars := []string{"ADMIN_URL", "ACCESS_KEY", "SECRET_KEY", "NATS_URL", "NATS_SUBJECT", "NODE_NAME", "INSTANCE_ID", "INTERVAL", "QUOTA_USAGE_PERCENT"}
	clearEnvVars(envVars)

	t.Run("uses defaults", func(t *testing.T) {
		cfg := quotausagemonitor.QuotaUsageMonitorConfig{
			AdminURL: "http://default:8080",
			Interval: 30,
		}
		result := mergeQuotaUsageMonitorConfigWithEnv(cfg)
		assert.Equal(t, "http://default:8080", result.AdminURL)
		assert.Equal(t, 30, result.Interval)
	})

	t.Run("overrides with env", func(t *testing.T) {
		os.Setenv("ADMIN_URL", "http://monitor:8080")
		os.Setenv("ACCESS_KEY", "monitor-key")
		os.Setenv("SECRET_KEY", "monitor-secret")
		os.Setenv("INTERVAL", "20")
		os.Setenv("QUOTA_USAGE_PERCENT", "75.0")
		defer clearEnvVars(envVars)

		cfg := quotausagemonitor.QuotaUsageMonitorConfig{}
		result := mergeQuotaUsageMonitorConfigWithEnv(cfg)

		assert.Equal(t, "http://monitor:8080", result.AdminURL)
		assert.Equal(t, "monitor-key", result.AccessKey)
		assert.Equal(t, "monitor-secret", result.SecretKey)
		assert.Equal(t, 20, result.Interval)
		assert.Equal(t, 75.0, result.QuotaUsagePercent)
	})
}
