// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package bucketnotify

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRGWNotification_JSONDeserialization(t *testing.T) {
	// Sample S3 notification JSON (based on AWS S3 event notification format)
	jsonData := `{
		"Records": [
			{
				"eventVersion": "2.1",
				"eventSource": "ceph:s3",
				"awsRegion": "us-east-1",
				"eventTime": "2024-01-15T12:00:00.000Z",
				"eventName": "s3:ObjectCreated:Put",
				"userIdentity": {
					"principalId": "user123"
				},
				"requestParameters": {
					"sourceIPAddress": "192.168.1.100"
				},
				"responseElements": {
					"x-amz-request-id": "req-12345",
					"x-amz-id-2": "id2-12345"
				},
				"s3": {
					"s3SchemaVersion": "1.0",
					"configurationId": "config123",
					"bucket": {
						"name": "my-bucket",
						"ownerIdentity": {
							"principalId": "owner123"
						},
						"arn": "arn:aws:s3:::my-bucket"
					},
					"object": {
						"key": "path/to/file.txt",
						"size": 1024,
						"eTag": "d41d8cd98f00b204e9800998ecf8427e",
						"versionId": "v1",
						"sequencer": "0A1B2C3D4E5F"
					}
				}
			}
		]
	}`

	var notification RGWNotification
	err := json.Unmarshal([]byte(jsonData), &notification)
	assert.NoError(t, err)

	// Verify records
	assert.Len(t, notification.Records, 1)

	record := notification.Records[0]
	assert.Equal(t, "2.1", record.EventVersion)
	assert.Equal(t, "ceph:s3", record.EventSource)
	assert.Equal(t, "us-east-1", record.AwsRegion)
	assert.Equal(t, "2024-01-15T12:00:00.000Z", record.EventTime)
	assert.Equal(t, "s3:ObjectCreated:Put", record.EventName)

	// User identity
	assert.Equal(t, "user123", record.UserIdentity.PrincipalID)

	// Request parameters
	assert.Equal(t, "192.168.1.100", record.RequestParameters.SourceIPAddress)

	// Response elements
	assert.Equal(t, "req-12345", record.ResponseElements.XAmzRequestID)
	assert.Equal(t, "id2-12345", record.ResponseElements.XAmzID2)

	// S3 details
	assert.Equal(t, "1.0", record.S3.S3SchemaVersion)
	assert.Equal(t, "config123", record.S3.ConfigurationID)

	// Bucket details
	assert.Equal(t, "my-bucket", record.S3.Bucket.Name)
	assert.Equal(t, "owner123", record.S3.Bucket.OwnerIdentity.PrincipalID)
	assert.Equal(t, "arn:aws:s3:::my-bucket", record.S3.Bucket.Arn)

	// Object details
	assert.Equal(t, "path/to/file.txt", record.S3.Object.Key)
	assert.Equal(t, int64(1024), record.S3.Object.Size)
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", record.S3.Object.ETag)
	assert.Equal(t, "v1", record.S3.Object.VersionID)
	assert.Equal(t, "0A1B2C3D4E5F", record.S3.Object.Sequencer)
}

func TestRGWNotification_MultipleRecords(t *testing.T) {
	jsonData := `{
		"Records": [
			{
				"eventName": "s3:ObjectCreated:Put",
				"s3": {
					"bucket": {"name": "bucket1"},
					"object": {"key": "file1.txt", "size": 100}
				}
			},
			{
				"eventName": "s3:ObjectRemoved:Delete",
				"s3": {
					"bucket": {"name": "bucket2"},
					"object": {"key": "file2.txt", "size": 200}
				}
			}
		]
	}`

	var notification RGWNotification
	err := json.Unmarshal([]byte(jsonData), &notification)
	assert.NoError(t, err)

	assert.Len(t, notification.Records, 2)
	assert.Equal(t, "s3:ObjectCreated:Put", notification.Records[0].EventName)
	assert.Equal(t, "bucket1", notification.Records[0].S3.Bucket.Name)
	assert.Equal(t, "s3:ObjectRemoved:Delete", notification.Records[1].EventName)
	assert.Equal(t, "bucket2", notification.Records[1].S3.Bucket.Name)
}

func TestRGWNotification_EmptyRecords(t *testing.T) {
	jsonData := `{"Records": []}`

	var notification RGWNotification
	err := json.Unmarshal([]byte(jsonData), &notification)
	assert.NoError(t, err)
	assert.Empty(t, notification.Records)
}

func TestRGWNotification_InvalidJSON(t *testing.T) {
	jsonData := `{invalid json}`

	var notification RGWNotification
	err := json.Unmarshal([]byte(jsonData), &notification)
	assert.Error(t, err)
}

func TestRGWNotification_EventTypes(t *testing.T) {
	eventTypes := []string{
		"s3:ObjectCreated:Put",
		"s3:ObjectCreated:Post",
		"s3:ObjectCreated:Copy",
		"s3:ObjectCreated:CompleteMultipartUpload",
		"s3:ObjectRemoved:Delete",
		"s3:ObjectRemoved:DeleteMarkerCreated",
	}

	for _, eventType := range eventTypes {
		t.Run(eventType, func(t *testing.T) {
			jsonData := `{"Records": [{"eventName": "` + eventType + `"}]}`

			var notification RGWNotification
			err := json.Unmarshal([]byte(jsonData), &notification)
			assert.NoError(t, err)
			assert.Equal(t, eventType, notification.Records[0].EventName)
		})
	}
}

func TestRGWNotification_LargeObjectSize(t *testing.T) {
	jsonData := `{
		"Records": [{
			"s3": {
				"object": {
					"key": "large-file.bin",
					"size": 5368709120
				}
			}
		}]
	}`

	var notification RGWNotification
	err := json.Unmarshal([]byte(jsonData), &notification)
	assert.NoError(t, err)
	assert.Equal(t, int64(5368709120), notification.Records[0].S3.Object.Size) // 5GB
}

func TestRGWNotification_SpecialCharactersInKey(t *testing.T) {
	jsonData := `{
		"Records": [{
			"s3": {
				"object": {
					"key": "path/to/file with spaces & special+chars.txt"
				}
			}
		}]
	}`

	var notification RGWNotification
	err := json.Unmarshal([]byte(jsonData), &notification)
	assert.NoError(t, err)
	assert.Equal(t, "path/to/file with spaces & special+chars.txt", notification.Records[0].S3.Object.Key)
}

func TestBucketNotifyConfig_StructFields(t *testing.T) {
	config := BucketNotifyConfig{
		EndpointPort: 8080,
		NatsURL:      "nats://localhost:4222",
		NatsSubject:  "rgw.notifications",
		UseNats:      true,
	}

	assert.Equal(t, 8080, config.EndpointPort)
	assert.Equal(t, "nats://localhost:4222", config.NatsURL)
	assert.Equal(t, "rgw.notifications", config.NatsSubject)
	assert.True(t, config.UseNats)
}

func TestBucketNotifyConfig_DefaultValues(t *testing.T) {
	config := BucketNotifyConfig{}

	assert.Equal(t, 0, config.EndpointPort)
	assert.Empty(t, config.NatsURL)
	assert.Empty(t, config.NatsSubject)
	assert.False(t, config.UseNats)
}

func TestBucketNotifyConfig_StandaloneMode(t *testing.T) {
	// Test configuration without NATS (prints to stdout)
	config := BucketNotifyConfig{
		EndpointPort: 9000,
		UseNats:      false,
	}

	assert.Equal(t, 9000, config.EndpointPort)
	assert.False(t, config.UseNats)
}

func TestBucketNotifyConfig_NatsEnabled(t *testing.T) {
	config := BucketNotifyConfig{
		EndpointPort: 8080,
		NatsURL:      "nats://nats-server:4222",
		NatsSubject:  "ceph.bucket.events",
		UseNats:      true,
	}

	assert.True(t, config.UseNats)
	assert.NotEmpty(t, config.NatsURL)
	assert.NotEmpty(t, config.NatsSubject)
}
