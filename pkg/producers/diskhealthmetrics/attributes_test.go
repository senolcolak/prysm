// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSMARTAttributes_DataIntegrity(t *testing.T) {
	// Verify the global SMARTAttributes slice has expected entries
	assert.NotEmpty(t, SMARTAttributes, "SMARTAttributes should not be empty")

	// Check a few critical attributes are present
	criticalIDs := map[int]bool{
		1:   true, // Raw_Read_Error_Rate
		5:   true, // Reallocated_Sector_Ct
		7:   true, // Seek_Error_Rate
		197: true, // Current_Pending_Sector
		198: true, // Offline_Uncorrectable
	}

	foundCritical := make(map[int]bool)
	for _, attr := range SMARTAttributes {
		if criticalIDs[attr.ID] {
			foundCritical[attr.ID] = true
			assert.True(t, attr.Critical, "Attribute ID %d should be marked as critical", attr.ID)
		}
	}

	for id := range criticalIDs {
		assert.True(t, foundCritical[id], "Critical attribute ID %d should be present", id)
	}
}

func TestSMARTAttributes_HasPromNames(t *testing.T) {
	// Verify each attribute has a Prometheus name
	for _, attr := range SMARTAttributes {
		assert.NotEmpty(t, attr.PromName, "Attribute ID %d (%s) should have PromName", attr.ID, attr.Key)
		assert.NotEmpty(t, attr.PromHelp, "Attribute ID %d (%s) should have PromHelp", attr.ID, attr.Key)
	}
}

func TestSMARTAttributes_UniqueIDs(t *testing.T) {
	// Verify all attribute IDs are unique
	seenIDs := make(map[int]bool)
	for _, attr := range SMARTAttributes {
		if seenIDs[attr.ID] {
			t.Errorf("Duplicate attribute ID: %d (%s)", attr.ID, attr.Key)
		}
		seenIDs[attr.ID] = true
	}
}

func TestSMARTAttributes_UniqueKeys(t *testing.T) {
	// Note: The SMARTAttributes list intentionally has some duplicate keys
	// because different SMART IDs can report similar metrics with different
	// interpretations. This test documents which keys appear multiple times.
	seenKeys := make(map[string]int)
	for _, attr := range SMARTAttributes {
		seenKeys[attr.Key]++
	}

	// These keys are known to appear multiple times (different SMART IDs, same name)
	expectedDuplicates := map[string]bool{
		"Soft_Read_Error_Rate":  true, // ID 13 and 201
		"G-Sense_Error_Rate":    true, // ID 191 and 221
		"Temperature_Celsius":   true, // ID 194 and 231
	}

	for key, count := range seenKeys {
		if count > 1 && !expectedDuplicates[key] {
			t.Errorf("Unexpected duplicate attribute key: %s (count: %d)", key, count)
		}
	}
}

func TestSMARTAttributes_TemperatureAttributes(t *testing.T) {
	// Find temperature-related attributes
	var tempAttrs []SMARTAttribute
	for _, attr := range SMARTAttributes {
		if attr.ID == 190 || attr.ID == 194 || attr.ID == 231 {
			tempAttrs = append(tempAttrs, attr)
		}
	}

	assert.NotEmpty(t, tempAttrs, "Should have temperature attributes")

	// Temperature attributes should not be critical
	for _, attr := range tempAttrs {
		assert.False(t, attr.Critical, "Temperature attribute %s (ID %d) should not be critical", attr.Key, attr.ID)
	}
}

func TestSMARTAttributes_SectorRelatedCritical(t *testing.T) {
	// Sector-related attributes should be critical
	sectorIDs := []int{5, 196, 197, 198} // Reallocated, Reallocation Event, Pending, Offline Uncorrectable

	for _, id := range sectorIDs {
		for _, attr := range SMARTAttributes {
			if attr.ID == id {
				assert.True(t, attr.Critical, "Sector attribute %s (ID %d) should be critical", attr.Key, attr.ID)
				break
			}
		}
	}
}

func TestSMARTAttribute_StructFields(t *testing.T) {
	attr := SMARTAttribute{
		ID:          5,
		Key:         "Reallocated_Sector_Ct",
		Name:        "Reallocated Sector Count",
		Value:       100,
		Critical:    true,
		Description: "Count of sectors moved to the spare area",
		PromName:    "disk_reallocated_sector_ct",
		PromHelp:    "Number of reallocated sectors on the disk",
	}

	assert.Equal(t, 5, attr.ID)
	assert.Equal(t, "Reallocated_Sector_Ct", attr.Key)
	assert.Equal(t, "Reallocated Sector Count", attr.Name)
	assert.Equal(t, uint64(100), attr.Value)
	assert.True(t, attr.Critical)
	assert.Equal(t, "Count of sectors moved to the spare area", attr.Description)
	assert.Equal(t, "disk_reallocated_sector_ct", attr.PromName)
	assert.Equal(t, "Number of reallocated sectors on the disk", attr.PromHelp)
}

func TestSMARTAttributes_PowerCycleNotCritical(t *testing.T) {
	// Power cycle and start/stop counts should not be critical
	nonCriticalIDs := []int{4, 12, 192, 193} // Start_Stop, Power_Cycle, Power-Off_Retract, Load_Unload

	for _, id := range nonCriticalIDs {
		for _, attr := range SMARTAttributes {
			if attr.ID == id {
				assert.False(t, attr.Critical, "Attribute %s (ID %d) should not be critical", attr.Key, attr.ID)
				break
			}
		}
	}
}

func TestSMARTAttributes_SpinRetryIsCritical(t *testing.T) {
	// Spin retry count should be critical
	for _, attr := range SMARTAttributes {
		if attr.ID == 10 {
			assert.True(t, attr.Critical, "Spin_Retry_Count should be critical")
			return
		}
	}
	t.Error("Spin_Retry_Count (ID 10) not found")
}
