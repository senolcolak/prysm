// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package radosgwusage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrysmStatus_UpdateTargetUp(t *testing.T) {
	tests := []struct {
		name     string
		up       bool
		expected float64
	}{
		{
			name:     "set target up",
			up:       true,
			expected: 1,
		},
		{
			name:     "set target down",
			up:       false,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &PrysmStatus{}
			status.UpdateTargetUp(tt.up)
			targetUp, _ := status.GetSnapshot()
			assert.Equal(t, tt.expected, targetUp)
		})
	}
}

func TestPrysmStatus_IncrementScrapeErrors(t *testing.T) {
	status := &PrysmStatus{}

	// Initially zero
	_, errors := status.GetSnapshot()
	assert.Equal(t, 0, errors)

	// Increment once
	status.IncrementScrapeErrors()
	_, errors = status.GetSnapshot()
	assert.Equal(t, 1, errors)

	// Increment multiple times
	status.IncrementScrapeErrors()
	status.IncrementScrapeErrors()
	_, errors = status.GetSnapshot()
	assert.Equal(t, 3, errors)
}

func TestPrysmStatus_GetSnapshot(t *testing.T) {
	status := &PrysmStatus{}

	// Initial state
	targetUp, errors := status.GetSnapshot()
	assert.Equal(t, float64(0), targetUp)
	assert.Equal(t, 0, errors)

	// After updates
	status.UpdateTargetUp(true)
	status.IncrementScrapeErrors()
	status.IncrementScrapeErrors()

	targetUp, errors = status.GetSnapshot()
	assert.Equal(t, float64(1), targetUp)
	assert.Equal(t, 2, errors)
}

func TestPrysmStatus_ConcurrentAccess(t *testing.T) {
	status := &PrysmStatus{}
	var wg sync.WaitGroup

	// Simulate concurrent access
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			status.UpdateTargetUp(true)
		}()
		go func() {
			defer wg.Done()
			status.IncrementScrapeErrors()
		}()
		go func() {
			defer wg.Done()
			status.GetSnapshot()
		}()
	}

	wg.Wait()

	// Should not panic and should have consistent state
	targetUp, errors := status.GetSnapshot()
	assert.Equal(t, float64(1), targetUp)
	assert.Equal(t, 100, errors) // 100 increments
}

func TestPrysmStatus_ToggleTargetUp(t *testing.T) {
	status := &PrysmStatus{}

	// Toggle up and down
	status.UpdateTargetUp(true)
	targetUp, _ := status.GetSnapshot()
	assert.Equal(t, float64(1), targetUp)

	status.UpdateTargetUp(false)
	targetUp, _ = status.GetSnapshot()
	assert.Equal(t, float64(0), targetUp)

	status.UpdateTargetUp(true)
	targetUp, _ = status.GetSnapshot()
	assert.Equal(t, float64(1), targetUp)
}
