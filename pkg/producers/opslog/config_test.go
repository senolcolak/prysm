// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package opslog

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsConfig_ApplyShortcuts_TrackEverything(t *testing.T) {
	cfg := &MetricsConfig{
		TrackEverything: true,
	}

	cfg.ApplyShortcuts()

	// Verify all detailed metrics are enabled
	assert.True(t, cfg.TrackRequestsDetailed)
	assert.True(t, cfg.TrackRequestsByMethodDetailed)
	assert.True(t, cfg.TrackRequestsByOperationDetailed)
	assert.True(t, cfg.TrackRequestsByStatusDetailed)
	assert.True(t, cfg.TrackBytesSentDetailed)
	assert.True(t, cfg.TrackBytesReceivedDetailed)
	assert.True(t, cfg.TrackErrorsDetailed)
	assert.True(t, cfg.TrackErrorsByIP)
	assert.True(t, cfg.TrackTimeoutErrors)
	assert.True(t, cfg.TrackErrorsByCategory)
	assert.True(t, cfg.TrackRequestsByIPDetailed)
	assert.True(t, cfg.TrackBytesSentByIPDetailed)
	assert.True(t, cfg.TrackBytesReceivedByIPDetailed)
	assert.True(t, cfg.TrackLatencyDetailed)
}

func TestMetricsConfig_ApplyShortcuts_NotEnabled(t *testing.T) {
	cfg := &MetricsConfig{
		TrackEverything: false,
	}

	cfg.ApplyShortcuts()

	// Verify nothing is enabled
	assert.False(t, cfg.TrackRequestsDetailed)
	assert.False(t, cfg.TrackRequestsByMethodDetailed)
	assert.False(t, cfg.TrackLatencyDetailed)
}

func TestMetricsConfig_DefaultValues(t *testing.T) {
	cfg := MetricsConfig{}

	// All booleans should default to false
	assert.False(t, cfg.TrackEverything)
	assert.False(t, cfg.TrackRequestsDetailed)
	assert.False(t, cfg.TrackRequestsPerUser)
	assert.False(t, cfg.TrackRequestsPerBucket)
	assert.False(t, cfg.TrackRequestsPerTenant)
	assert.False(t, cfg.TrackLatencyDetailed)
}

func TestOpsLogConfig_StructFields(t *testing.T) {
	cfg := OpsLogConfig{
		LogFilePath:               "/var/log/ceph/ops.log",
		TruncateLogOnStart:        true,
		SocketPath:                "/var/run/prysm.sock",
		NatsURL:                   "nats://localhost:4222",
		NatsSubject:               "rgw.ops",
		NatsMetricsSubject:        "rgw.ops.metrics",
		UseNats:                   true,
		LogToStdout:               false,
		LogPrettyPrint:            true,
		LogRetentionDays:          7,
		MaxLogFileSize:            104857600, // 100MB
		Prometheus:                true,
		PrometheusPort:            9090,
		PodName:                   "rgw-pod-1",
		IgnoreAnonymousRequests:   true,
		PrometheusIntervalSeconds: 60,
	}

	assert.Equal(t, "/var/log/ceph/ops.log", cfg.LogFilePath)
	assert.True(t, cfg.TruncateLogOnStart)
	assert.Equal(t, "/var/run/prysm.sock", cfg.SocketPath)
	assert.Equal(t, "nats://localhost:4222", cfg.NatsURL)
	assert.Equal(t, "rgw.ops", cfg.NatsSubject)
	assert.Equal(t, "rgw.ops.metrics", cfg.NatsMetricsSubject)
	assert.True(t, cfg.UseNats)
	assert.False(t, cfg.LogToStdout)
	assert.True(t, cfg.LogPrettyPrint)
	assert.Equal(t, 7, cfg.LogRetentionDays)
	assert.Equal(t, int64(104857600), cfg.MaxLogFileSize)
	assert.True(t, cfg.Prometheus)
	assert.Equal(t, 9090, cfg.PrometheusPort)
	assert.Equal(t, "rgw-pod-1", cfg.PodName)
	assert.True(t, cfg.IgnoreAnonymousRequests)
	assert.Equal(t, 60, cfg.PrometheusIntervalSeconds)
}

func TestAuditSinkConfig_StructFields(t *testing.T) {
	cfg := AuditSinkConfig{
		Enabled:           true,
		RabbitMQURL:       "amqp://guest:guest@localhost:5672/",
		QueueName:         "audit-events",
		InternalQueueSize: 100,
		Debug:             true,
	}

	assert.True(t, cfg.Enabled)
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.RabbitMQURL)
	assert.Equal(t, "audit-events", cfg.QueueName)
	assert.Equal(t, 100, cfg.InternalQueueSize)
	assert.True(t, cfg.Debug)
}

func TestAuditSinkConfig_DefaultValues(t *testing.T) {
	cfg := AuditSinkConfig{}

	assert.False(t, cfg.Enabled)
	assert.Empty(t, cfg.RabbitMQURL)
	assert.Empty(t, cfg.QueueName)
	assert.Equal(t, 0, cfg.InternalQueueSize)
	assert.False(t, cfg.Debug)
}

func TestS3OperationLog_StructFields(t *testing.T) {
	log := S3OperationLog{
		Bucket:             "my-bucket",
		Object:             "path/to/file.txt",
		Time:               "2024-01-15T12:30:45.123456Z",
		TimeLocal:          "15/Jan/2024:12:30:45 +0000",
		RemoteAddr:         "192.168.1.100",
		User:               "tenant$user",
		Operation:          "get_obj",
		URI:                "GET /my-bucket/path/to/file.txt HTTP/1.1",
		HTTPStatus:         "200",
		ErrorCode:          "",
		BytesSent:          1024,
		BytesReceived:      0,
		ObjectSize:         1024,
		TotalTime:          150,
		UserAgent:          "aws-cli/2.0",
		Referrer:           "https://example.com",
		TransID:            "tx123456",
		AuthenticationType: "keystone",
		AccessKeyID:        "AKIAIOSFODNN7EXAMPLE",
		TempURL:            false,
	}

	assert.Equal(t, "my-bucket", log.Bucket)
	assert.Equal(t, "path/to/file.txt", log.Object)
	assert.Equal(t, "2024-01-15T12:30:45.123456Z", log.Time)
	assert.Equal(t, "192.168.1.100", log.RemoteAddr)
	assert.Equal(t, "tenant$user", log.User)
	assert.Equal(t, "get_obj", log.Operation)
	assert.Equal(t, "200", log.HTTPStatus)
	assert.Equal(t, 1024, log.BytesSent)
	assert.Equal(t, 0, log.BytesReceived)
	assert.Equal(t, 150, log.TotalTime)
	assert.Equal(t, "aws-cli/2.0", log.UserAgent)
	assert.Equal(t, "tx123456", log.TransID)
	assert.Equal(t, "keystone", log.AuthenticationType)
	assert.False(t, log.TempURL)
}

func TestS3OperationLog_JSONSerialization(t *testing.T) {
	log := S3OperationLog{
		Bucket:     "test-bucket",
		Object:     "test.txt",
		User:       "user123",
		Operation:  "put_obj",
		HTTPStatus: "201",
		BytesSent:  512,
	}

	jsonData, err := json.Marshal(log)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Equal(t, "test-bucket", result["bucket"])
	assert.Equal(t, "test.txt", result["object"])
	assert.Equal(t, "user123", result["user"])
	assert.Equal(t, "put_obj", result["operation"])
	assert.Equal(t, "201", result["http_status"])
	assert.Equal(t, float64(512), result["bytes_sent"])
}

func TestS3OperationLog_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"bucket": "my-bucket",
		"object": "file.txt",
		"time": "2024-01-15T12:30:45.123456Z",
		"remote_addr": "10.0.0.1",
		"user": "project$user",
		"operation": "get_obj",
		"uri": "GET /my-bucket/file.txt HTTP/1.1",
		"http_status": "200",
		"bytes_sent": 2048,
		"bytes_received": 0,
		"total_time": 100,
		"user_agent": "curl/7.68.0",
		"trans_id": "tx999"
	}`

	var log S3OperationLog
	err := json.Unmarshal([]byte(jsonData), &log)
	assert.NoError(t, err)

	assert.Equal(t, "my-bucket", log.Bucket)
	assert.Equal(t, "file.txt", log.Object)
	assert.Equal(t, "10.0.0.1", log.RemoteAddr)
	assert.Equal(t, "project$user", log.User)
	assert.Equal(t, "get_obj", log.Operation)
	assert.Equal(t, "200", log.HTTPStatus)
	assert.Equal(t, 2048, log.BytesSent)
	assert.Equal(t, 100, log.TotalTime)
}

func TestS3OperationLog_CleanupBucketName(t *testing.T) {
	tests := []struct {
		name     string
		bucket   string
		expected string
	}{
		{
			name:     "simple bucket name",
			bucket:   "my-bucket",
			expected: "my-bucket",
		},
		{
			name:     "bucket with tenant prefix",
			bucket:   "tenant/my-bucket",
			expected: "my-bucket",
		},
		{
			name:     "bucket with multiple slashes",
			bucket:   "tenant/user/my-bucket",
			expected: "my-bucket",
		},
		{
			name:     "empty bucket",
			bucket:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &S3OperationLog{Bucket: tt.bucket}
			log.CleanupBucketName()
			assert.Equal(t, tt.expected, log.Bucket)
		})
	}
}

func TestExtractUserAndTenant(t *testing.T) {
	tests := []struct {
		name           string
		user           string
		expectedUser   string
		expectedTenant string
	}{
		{
			name:           "user with tenant",
			user:           "tenant$user",
			expectedUser:   "tenant",
			expectedTenant: "user",
		},
		{
			name:           "user without tenant",
			user:           "simpleuser",
			expectedUser:   "simpleuser",
			expectedTenant: "none",
		},
		{
			name:           "empty user",
			user:           "",
			expectedUser:   "",
			expectedTenant: "none",
		},
		{
			name:           "multiple dollar signs",
			user:           "a$b$c",
			expectedUser:   "a",
			expectedTenant: "b$c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, tenant := extractUserAndTenant(tt.user)
			assert.Equal(t, tt.expectedUser, user)
			assert.Equal(t, tt.expectedTenant, tenant)
		})
	}
}

func TestKeystoneScope_StructFields(t *testing.T) {
	scope := KeystoneScope{
		Project: KeystoneProject{
			ID:   "proj-123",
			Name: "my-project",
			Domain: KeystoneDomain{
				ID:   "domain-123",
				Name: "my-domain",
			},
		},
		User: KeystoneUser{
			ID:   "user-123",
			Name: "john.doe",
			Domain: KeystoneDomain{
				ID:   "user-domain-123",
				Name: "user-domain",
			},
		},
		Roles: []string{"admin", "member"},
		ApplicationCredential: &KeystoneApplicationCredential{
			ID:         "app-123",
			Name:       "my-app-cred",
			Restricted: false,
		},
	}

	assert.Equal(t, "proj-123", scope.Project.ID)
	assert.Equal(t, "my-project", scope.Project.Name)
	assert.Equal(t, "domain-123", scope.Project.Domain.ID)
	assert.Equal(t, "user-123", scope.User.ID)
	assert.Equal(t, "john.doe", scope.User.Name)
	assert.Equal(t, []string{"admin", "member"}, scope.Roles)
	assert.Equal(t, "app-123", scope.ApplicationCredential.ID)
	assert.Equal(t, "my-app-cred", scope.ApplicationCredential.Name)
	assert.False(t, scope.ApplicationCredential.Restricted)
}
