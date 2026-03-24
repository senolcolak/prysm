package opslog

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubtractMetrics(t *testing.T) {
	total := NewMetrics()
	prev := NewMetrics()

	// Setup total values across different metric types
	total.TotalRequests.Store(10)
	total.BytesSent.Store(2048)

	// Test various storage maps
	total.RequestsDetailed.Store("user1|bucket1|GET|200", newUint64(5))
	total.RequestsByMethodDetailed.Store("user1|bucket1|GET", newUint64(3))
	total.BytesSentDetailed.Store("user1|bucket1", newUint64(1024))
	total.ErrorsDetailed.Store("user1|bucket1|404", newUint64(2))

	// Setup previous values
	prev.TotalRequests.Store(7)
	prev.BytesSent.Store(1024)
	prev.RequestsDetailed.Store("user1|bucket1|GET|200", newUint64(2))
	prev.BytesSentDetailed.Store("user1|bucket1", newUint64(512))

	// Subtract
	delta := SubtractMetrics(total, prev)

	// Test atomic counters
	assert.Equal(t, uint64(3), delta.TotalRequests.Load())
	assert.Equal(t, uint64(1024), delta.BytesSent.Load())

	// Test detailed requests delta
	v1, ok := delta.RequestsDetailed.Load("user1|bucket1|GET|200")
	assert.True(t, ok, "Expected key user1|bucket1|GET|200 to exist in RequestsDetailed")
	assert.Equal(t, uint64(3), v1.(*atomic.Uint64).Load())

	// Test method details (new key, should equal total)
	v2, ok := delta.RequestsByMethodDetailed.Load("user1|bucket1|GET")
	assert.True(t, ok, "Expected key user1|bucket1|GET to exist in RequestsByMethodDetailed")
	assert.Equal(t, uint64(3), v2.(*atomic.Uint64).Load())

	// Test bytes delta
	v3, ok := delta.BytesSentDetailed.Load("user1|bucket1")
	assert.True(t, ok, "Expected key user1|bucket1 to exist in BytesSentDetailed")
	assert.Equal(t, uint64(512), v3.(*atomic.Uint64).Load())

	// Test errors (new key, should equal total)
	v4, ok := delta.ErrorsDetailed.Load("user1|bucket1|404")
	assert.True(t, ok, "Expected key user1|bucket1|404 to exist in ErrorsDetailed")
	assert.Equal(t, uint64(2), v4.(*atomic.Uint64).Load())
}

func TestCloneMetrics(t *testing.T) {
	original := NewMetrics()

	// Set some base values across different metric types
	original.TotalRequests.Store(42)
	original.BytesSent.Store(1024)
	original.Errors.Store(5)

	// Test different storage maps
	original.RequestsDetailed.Store("user1|bucket1|GET|200", newUint64(5))
	original.RequestsByMethodDetailed.Store("user1|bucket1|GET", newUint64(3))
	original.BytesSentDetailed.Store("user1|bucket1", newUint64(1024))
	original.BytesSentPerUser.Store("user1", newUint64(888))
	original.ErrorsDetailed.Store("user1|bucket1|404", newUint64(2))
	original.RequestsByIPDetailed.Store("user1|192.168.1.1", newUint64(7))

	// Clone it
	clone := original.Clone()

	// Test top-level atomic fields
	assert.Equal(t, uint64(42), clone.TotalRequests.Load())
	assert.Equal(t, uint64(1024), clone.BytesSent.Load())
	assert.Equal(t, uint64(5), clone.Errors.Load())

	// Test sync.Map values across different types
	v1, ok := clone.RequestsDetailed.Load("user1|bucket1|GET|200")
	assert.True(t, ok, "Expected key to exist in RequestsDetailed")
	assert.Equal(t, uint64(5), v1.(*atomic.Uint64).Load())

	v2, ok := clone.RequestsByMethodDetailed.Load("user1|bucket1|GET")
	assert.True(t, ok, "Expected key to exist in RequestsByMethodDetailed")
	assert.Equal(t, uint64(3), v2.(*atomic.Uint64).Load())

	v3, ok := clone.BytesSentDetailed.Load("user1|bucket1")
	assert.True(t, ok, "Expected key to exist in BytesSentDetailed")
	assert.Equal(t, uint64(1024), v3.(*atomic.Uint64).Load())

	v4, ok := clone.BytesSentPerUser.Load("user1")
	assert.True(t, ok, "Expected key to exist in BytesSentPerUser")
	assert.Equal(t, uint64(888), v4.(*atomic.Uint64).Load())

	v5, ok := clone.ErrorsDetailed.Load("user1|bucket1|404")
	assert.True(t, ok, "Expected key to exist in ErrorsDetailed")
	assert.Equal(t, uint64(2), v5.(*atomic.Uint64).Load())

	v6, ok := clone.RequestsByIPDetailed.Load("user1|192.168.1.1")
	assert.True(t, ok, "Expected key to exist in RequestsByIPDetailed")
	assert.Equal(t, uint64(7), v6.(*atomic.Uint64).Load())

	// Mutate original, ensure clone is untouched
	original.TotalRequests.Add(10)
	original.RequestsDetailed.Store("user1|bucket1|GET|200", newUint64(99))

	// Verify clone remains unchanged
	assert.Equal(t, uint64(42), clone.TotalRequests.Load(), "Clone TotalRequests should remain unchanged")

	v1After, _ := clone.RequestsDetailed.Load("user1|bucket1|GET|200")
	assert.Equal(t, uint64(5), v1After.(*atomic.Uint64).Load(), "Clone RequestsDetailed should remain unchanged")
}

func TestSubtractMetrics_ZeroDelta(t *testing.T) {
	total := NewMetrics()
	prev := NewMetrics()

	// Test zero delta across different metric types
	total.RequestsDetailed.Store("user1|bucket1|GET|200", newUint64(5))
	prev.RequestsDetailed.Store("user1|bucket1|GET|200", newUint64(5))

	total.BytesSentPerUser.Store("user1", newUint64(1024))
	prev.BytesSentPerUser.Store("user1", newUint64(1024))

	delta := SubtractMetrics(total, prev)

	// Zero deltas should not be stored
	_, ok1 := delta.RequestsDetailed.Load("user1|bucket1|GET|200")
	assert.False(t, ok1, "Zero delta should not be stored in RequestsDetailed")

	_, ok2 := delta.BytesSentPerUser.Load("user1")
	assert.False(t, ok2, "Zero delta should not be stored in BytesSentPerUser")
}

func TestSubtractMetrics_MissingInPrev(t *testing.T) {
	total := NewMetrics()
	prev := NewMetrics()

	// Test new keys that don't exist in previous
	total.RequestsDetailed.Store("new|key|GET|200", newUint64(7))
	total.ErrorsPerUser.Store("newuser|404", newUint64(3))

	delta := SubtractMetrics(total, prev)

	// New keys should appear with full value
	v1, ok1 := delta.RequestsDetailed.Load("new|key|GET|200")
	assert.True(t, ok1, "New key should exist in delta")
	assert.Equal(t, uint64(7), v1.(*atomic.Uint64).Load())

	v2, ok2 := delta.ErrorsPerUser.Load("newuser|404")
	assert.True(t, ok2, "New error key should exist in delta")
	assert.Equal(t, uint64(3), v2.(*atomic.Uint64).Load())
}

func TestLatencyObsPropagation(t *testing.T) {
	called := false
	callCount := 0
	var capturedArgs []string

	cb := func(u, tnt, bucket, method string, sec float64) {
		called = true
		callCount++
		capturedArgs = []string{u, tnt, bucket, method}
		assert.Equal(t, "u1", u)
		assert.Equal(t, "t1", tnt)
		assert.Equal(t, "b1", bucket)
		assert.Equal(t, "M", method)
		assert.InDelta(t, 0.123, sec, 1e-6)
	}

	// Test direct call
	m := NewMetrics(cb)
	m.LatencyObs("u1", "t1", "b1", "M", 0.123)
	assert.True(t, called, "LatencyObs should be called")
	assert.Equal(t, 1, callCount)
	assert.Equal(t, []string{"u1", "t1", "b1", "M"}, capturedArgs)

	// Test clone carries it forward
	clone := m.Clone()
	called = false
	callCount = 0
	clone.LatencyObs("u1", "t1", "b1", "M", 0.123)
	assert.True(t, called, "Cloned LatencyObs should be called")
	assert.Equal(t, 1, callCount)

	// Test subtract carries it forward
	total := NewMetrics(cb)
	prev := NewMetrics(cb)
	delta := SubtractMetrics(total, prev)
	called = false
	callCount = 0
	delta.LatencyObs("u1", "t1", "b1", "M", 0.123)
	assert.True(t, called, "Delta LatencyObs should be called")
	assert.Equal(t, 1, callCount)
}

func TestLatencyObsDefaultNoOp(t *testing.T) {
	// Test that NewMetrics without callback creates no-op function
	m := NewMetrics()

	// Should not panic
	assert.NotPanics(t, func() {
		m.LatencyObs("user", "tenant", "bucket", "method", 1.23)
	}, "Default LatencyObs should be no-op and not panic")
}

func TestMetricsUpdate_BasicFunctionality(t *testing.T) {
	config := &MetricsConfig{
		TrackRequestsDetailed:  true,
		TrackLatencyDetailed:   true,
		TrackBytesSentDetailed: true,
		TrackErrorsDetailed:    true,
	}

	latencyCallCount := 0
	latencyObs := func(user, tenant, bucket, method string, seconds float64) {
		latencyCallCount++
		assert.Equal(t, "user1", user)
		assert.Equal(t, "tenant1", tenant)
		assert.Equal(t, "bucket1", bucket)
		assert.Equal(t, "GET", method)
		assert.InDelta(t, 0.150, seconds, 1e-6) // 150ms converted to seconds
	}

	m := NewMetrics(latencyObs)

	logEntry := S3OperationLog{
		User:          "user1$tenant1",
		Bucket:        "bucket1",
		URI:           "GET /bucket1/object.txt HTTP/1.1",
		HTTPStatus:    "200",
		BytesSent:     1024,
		BytesReceived: 0,
		TotalTime:     150, // milliseconds
	}

	// Update metrics
	m.Update(logEntry, config)

	// Verify atomic counters
	assert.Equal(t, uint64(1), m.TotalRequests.Load())
	assert.Equal(t, uint64(1024), m.BytesSent.Load())
	assert.Equal(t, uint64(0), m.BytesReceived.Load())
	assert.Equal(t, uint64(0), m.Errors.Load()) // 200 is not an error

	// Verify detailed requests tracking
	v, ok := m.RequestsDetailed.Load("user1$tenant1|bucket1|GET|200")
	assert.True(t, ok, "Should track detailed request")
	assert.Equal(t, uint64(1), v.(*atomic.Uint64).Load())

	// Verify bytes tracking
	v2, ok2 := m.BytesSentDetailed.Load("user1$tenant1|bucket1")
	assert.True(t, ok2, "Should track detailed bytes sent")
	assert.Equal(t, uint64(1024), v2.(*atomic.Uint64).Load())

	// Verify latency observation was called
	assert.Equal(t, 1, latencyCallCount, "LatencyObs should be called once")
}

func TestMetricsUpdate_ErrorTracking(t *testing.T) {
	config := &MetricsConfig{
		TrackErrorsDetailed: true,
		TrackErrorsPerUser:  true,
	}

	m := NewMetrics()

	logEntry := S3OperationLog{
		User:       "user1$tenant1",
		Bucket:     "bucket1",
		URI:        "GET /bucket1/missing.txt HTTP/1.1",
		HTTPStatus: "404",
	}

	m.Update(logEntry, config)

	// Verify error counters
	assert.Equal(t, uint64(1), m.Errors.Load())

	// Verify detailed error tracking
	v1, ok1 := m.ErrorsDetailed.Load("user1$tenant1|bucket1|404")
	assert.True(t, ok1, "Should track detailed error")
	assert.Equal(t, uint64(1), v1.(*atomic.Uint64).Load())

	// Verify per-user error tracking
	v2, ok2 := m.ErrorsPerUser.Load("user1|404")
	assert.True(t, ok2, "Should track per-user error")
	assert.Equal(t, uint64(1), v2.(*atomic.Uint64).Load())
}

func TestMetricsUpdate_ConditionalTracking(t *testing.T) {
	// Test that disabled tracking doesn't create entries
	config := &MetricsConfig{
		TrackRequestsDetailed: false,
		TrackErrorsDetailed:   false,
		TrackLatencyDetailed:  false,
	}

	m := NewMetrics()

	logEntry := S3OperationLog{
		User:       "user1$tenant1",
		Bucket:     "bucket1",
		URI:        "GET /bucket1/object.txt HTTP/1.1",
		HTTPStatus: "404",
		TotalTime:  150,
	}

	m.Update(logEntry, config)

	// Basic counters should still work
	assert.Equal(t, uint64(1), m.TotalRequests.Load())
	assert.Equal(t, uint64(1), m.Errors.Load())

	// But detailed tracking should be empty
	_, ok1 := m.RequestsDetailed.Load("user1$tenant1|bucket1|GET|404")
	assert.False(t, ok1, "Should not track detailed requests when disabled")

	_, ok2 := m.ErrorsDetailed.Load("user1$tenant1|bucket1|404")
	assert.False(t, ok2, "Should not track detailed errors when disabled")
}

func newUint64(val uint64) *atomic.Uint64 {
	var u atomic.Uint64
	u.Store(val)
	return &u
}

func TestExtractHTTPMethod(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "GET request",
			uri:      "GET /bucket/object HTTP/1.1",
			expected: "GET",
		},
		{
			name:     "PUT request",
			uri:      "PUT /bucket/newfile.txt HTTP/1.1",
			expected: "PUT",
		},
		{
			name:     "POST request",
			uri:      "POST /bucket?uploads HTTP/1.1",
			expected: "POST",
		},
		{
			name:     "DELETE request",
			uri:      "DELETE /bucket/file.txt HTTP/1.1",
			expected: "DELETE",
		},
		{
			name:     "HEAD request",
			uri:      "HEAD /bucket HTTP/1.1",
			expected: "HEAD",
		},
		{
			name:     "OPTIONS request",
			uri:      "OPTIONS /bucket HTTP/1.1",
			expected: "OPTIONS",
		},
		{
			name:     "PATCH request",
			uri:      "PATCH /bucket/object HTTP/1.1",
			expected: "PATCH",
		},
		{
			name:     "empty URI",
			uri:      "",
			expected: "UNKNOWN",
		},
		{
			name:     "invalid method",
			uri:      "INVALID /bucket HTTP/1.1",
			expected: "UNKNOWN",
		},
		{
			name:     "only path",
			uri:      "/bucket/object",
			expected: "UNKNOWN",
		},
		{
			name:     "lowercase method",
			uri:      "get /bucket HTTP/1.1",
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractHTTPMethod(tt.uri)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := NewMetrics()

	// Add some data
	m.TotalRequests.Add(100)
	m.BytesSent.Add(1000)
	m.BytesReceived.Add(500)
	m.Errors.Add(5)

	m.RequestsDetailed.Store("user|bucket|GET|200", newUint64(10))
	m.RequestsByMethodGlobal.Store("GET", newUint64(20))
	m.BytesSentPerUser.Store("user1", newUint64(1024))
	m.ErrorsPerStatus.Store("404", newUint64(5))

	// Verify data exists
	assert.Equal(t, uint64(100), m.TotalRequests.Load())
	v, ok := m.RequestsDetailed.Load("user|bucket|GET|200")
	assert.True(t, ok)
	assert.Equal(t, uint64(10), v.(*atomic.Uint64).Load())

	// Reset
	m.Reset()

	// Verify all counters are zero
	assert.Equal(t, uint64(0), m.TotalRequests.Load())
	assert.Equal(t, uint64(0), m.BytesSent.Load())
	assert.Equal(t, uint64(0), m.BytesReceived.Load())
	assert.Equal(t, uint64(0), m.Errors.Load())

	// Verify maps are empty
	_, ok = m.RequestsDetailed.Load("user|bucket|GET|200")
	assert.False(t, ok, "RequestsDetailed should be empty after reset")

	_, ok = m.RequestsByMethodGlobal.Load("GET")
	assert.False(t, ok, "RequestsByMethodGlobal should be empty after reset")

	_, ok = m.BytesSentPerUser.Load("user1")
	assert.False(t, ok, "BytesSentPerUser should be empty after reset")

	_, ok = m.ErrorsPerStatus.Load("404")
	assert.False(t, ok, "ErrorsPerStatus should be empty after reset")
}

func TestMetrics_ToJSON_Minimal(t *testing.T) {
	m := NewMetrics()
	m.TotalRequests.Add(100)
	m.BytesSent.Add(1048576)
	m.BytesReceived.Add(524288)
	m.Errors.Add(5)

	config := &MetricsConfig{}

	jsonData, err := m.ToJSON(config)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	assert.Equal(t, float64(100), result["total_requests"])
	assert.Equal(t, float64(1048576), result["bytes_sent"])
	assert.Equal(t, float64(524288), result["bytes_received"])
	assert.Equal(t, float64(5), result["errors"])

	// Detailed fields should not be present
	_, hasDetailed := result["requests_detailed"]
	assert.False(t, hasDetailed, "requests_detailed should not be present when tracking disabled")
}

func TestMetrics_ToJSON_WithAllTracking(t *testing.T) {
	m := NewMetrics()
	m.TotalRequests.Add(10)

	m.RequestsDetailed.Store("user|bucket|GET|200", newUint64(5))
	m.RequestsByUser.Store("user|bucket|GET|200", newUint64(3))
	m.BytesSentDetailed.Store("user|bucket", newUint64(1024))
	m.ErrorsDetailed.Store("user|bucket|404", newUint64(2))

	config := &MetricsConfig{
		TrackRequestsDetailed:  true,
		TrackRequestsPerUser:   true,
		TrackBytesSentDetailed: true,
		TrackErrorsDetailed:    true,
	}

	jsonData, err := m.ToJSON(config)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verify detailed fields are present
	detailed, ok := result["requests_detailed"].(map[string]interface{})
	assert.True(t, ok, "requests_detailed should be a map")
	assert.Equal(t, float64(5), detailed["user|bucket|GET|200"])

	byUser, ok := result["requests_by_user"].(map[string]interface{})
	assert.True(t, ok, "requests_by_user should be a map")
	assert.Equal(t, float64(3), byUser["user|bucket|GET|200"])

	bytesSent, ok := result["bytes_sent_detailed"].(map[string]interface{})
	assert.True(t, ok, "bytes_sent_detailed should be a map")
	assert.Equal(t, float64(1024), bytesSent["user|bucket"])
}

func TestDiff(t *testing.T) {
	tests := []struct {
		name     string
		a        uint64
		b        uint64
		expected uint64
	}{
		{
			name:     "a greater than b",
			a:        100,
			b:        30,
			expected: 70,
		},
		{
			name:     "a equals b",
			a:        50,
			b:        50,
			expected: 0,
		},
		{
			name:     "a less than b (no underflow)",
			a:        30,
			b:        100,
			expected: 0,
		},
		{
			name:     "both zero",
			a:        0,
			b:        0,
			expected: 0,
		},
		{
			name:     "large numbers",
			a:        1000000000,
			b:        999999999,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diff(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIncrementSyncMap(t *testing.T) {
	var m sync.Map

	// First increment creates the key
	incrementSyncMap(&m, "key1")
	v, ok := m.Load("key1")
	assert.True(t, ok)
	assert.Equal(t, uint64(1), v.(*atomic.Uint64).Load())

	// Subsequent increments increase the value
	incrementSyncMap(&m, "key1")
	incrementSyncMap(&m, "key1")
	v, _ = m.Load("key1")
	assert.Equal(t, uint64(3), v.(*atomic.Uint64).Load())

	// Different key
	incrementSyncMap(&m, "key2")
	v, _ = m.Load("key2")
	assert.Equal(t, uint64(1), v.(*atomic.Uint64).Load())
}

func TestIncrementSyncMapValue(t *testing.T) {
	var m sync.Map

	// First increment with value
	incrementSyncMapValue(&m, "bytes_key", 1000)
	v, ok := m.Load("bytes_key")
	assert.True(t, ok)
	assert.Equal(t, uint64(1000), v.(*atomic.Uint64).Load())

	// Add more
	incrementSyncMapValue(&m, "bytes_key", 500)
	v, _ = m.Load("bytes_key")
	assert.Equal(t, uint64(1500), v.(*atomic.Uint64).Load())

	// Large value
	incrementSyncMapValue(&m, "large_key", 1073741824)
	v, _ = m.Load("large_key")
	assert.Equal(t, uint64(1073741824), v.(*atomic.Uint64).Load())
}

func TestLoadSyncMap(t *testing.T) {
	var m sync.Map

	m.Store("key1", newUint64(100))
	m.Store("key2", newUint64(200))
	m.Store("key3", newUint64(300))

	result := loadSyncMap(&m)

	assert.Equal(t, uint64(100), result["key1"])
	assert.Equal(t, uint64(200), result["key2"])
	assert.Equal(t, uint64(300), result["key3"])
	assert.Len(t, result, 3)
}

func TestResetSyncMap(t *testing.T) {
	var m sync.Map

	m.Store("key1", newUint64(1))
	m.Store("key2", newUint64(2))
	m.Store("key3", newUint64(3))

	// Verify keys exist
	result := loadSyncMap(&m)
	assert.Len(t, result, 3)

	// Reset
	resetSyncMap(&m)

	// Verify empty
	result = loadSyncMap(&m)
	assert.Empty(t, result)
}

func TestCopySyncMap(t *testing.T) {
	var src, dst sync.Map

	src.Store("key1", newUint64(100))
	src.Store("key2", newUint64(200))

	copySyncMap(&src, &dst)

	// Verify values are copied
	v1, ok1 := dst.Load("key1")
	assert.True(t, ok1)
	assert.Equal(t, uint64(100), v1.(*atomic.Uint64).Load())

	v2, ok2 := dst.Load("key2")
	assert.True(t, ok2)
	assert.Equal(t, uint64(200), v2.(*atomic.Uint64).Load())

	// Verify independence - modifying source doesn't affect dest
	if val, ok := src.Load("key1"); ok {
		val.(*atomic.Uint64).Add(50)
	}

	v1After, _ := dst.Load("key1")
	assert.Equal(t, uint64(100), v1After.(*atomic.Uint64).Load(), "Dest should be unchanged")
}

func TestUpdateMaxAtomic(t *testing.T) {
	var target atomic.Uint64

	// First update sets the value
	updateMaxAtomic(&target, 50)
	assert.Equal(t, uint64(50), target.Load())

	// Higher value updates
	updateMaxAtomic(&target, 100)
	assert.Equal(t, uint64(100), target.Load())

	// Lower value doesn't update
	updateMaxAtomic(&target, 30)
	assert.Equal(t, uint64(100), target.Load())

	// Equal value stays same
	updateMaxAtomic(&target, 100)
	assert.Equal(t, uint64(100), target.Load())
}

func TestUpdateMinAtomic(t *testing.T) {
	var target atomic.Uint64

	// First update with zero initial value
	updateMinAtomic(&target, 50)
	assert.Equal(t, uint64(50), target.Load())

	// Lower value updates
	updateMinAtomic(&target, 30)
	assert.Equal(t, uint64(30), target.Load())

	// Higher value doesn't update
	updateMinAtomic(&target, 100)
	assert.Equal(t, uint64(30), target.Load())
}

func TestUpdateMaxSyncMap(t *testing.T) {
	var m sync.Map

	updateMaxSyncMap(&m, "max_key", 50)
	v, _ := m.Load("max_key")
	assert.Equal(t, uint64(50), v.(*atomic.Uint64).Load())

	updateMaxSyncMap(&m, "max_key", 100)
	v, _ = m.Load("max_key")
	assert.Equal(t, uint64(100), v.(*atomic.Uint64).Load())

	updateMaxSyncMap(&m, "max_key", 30)
	v, _ = m.Load("max_key")
	assert.Equal(t, uint64(100), v.(*atomic.Uint64).Load()) // Unchanged
}

func TestUpdateMinSyncMap(t *testing.T) {
	var m sync.Map

	updateMinSyncMap(&m, "min_key", 50)
	v, _ := m.Load("min_key")
	assert.Equal(t, uint64(50), v.(*atomic.Uint64).Load())

	updateMinSyncMap(&m, "min_key", 30)
	v, _ = m.Load("min_key")
	assert.Equal(t, uint64(30), v.(*atomic.Uint64).Load())

	updateMinSyncMap(&m, "min_key", 100)
	v, _ = m.Load("min_key")
	assert.Equal(t, uint64(30), v.(*atomic.Uint64).Load()) // Unchanged
}
