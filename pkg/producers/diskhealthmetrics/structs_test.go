// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSmartAttributes_ReturnsExpectedKeys(t *testing.T) {
	attrs := GetSmartAttributes()

	// Verify some essential attributes exist
	expectedKeys := []string{
		"temperature_celsius",
		"power_on_hours",
		"reallocated_sector_ct",
		"current_pending_sector",
		"power_cycle_count",
		"raw_read_error_rate",
		"percentage_used",
		"available_spare",
		"unsafe_shutdowns",
	}

	for _, key := range expectedKeys {
		_, exists := attrs[key]
		assert.True(t, exists, "Expected attribute %q to exist", key)
	}
}

func TestGetSmartAttributes_DefaultValues(t *testing.T) {
	attrs := GetSmartAttributes()

	// All attributes should start with -1 values (unset)
	for key, attr := range attrs {
		assert.Equal(t, int64(-1), attr.Threshold, "Attribute %s threshold should be -1", key)
		assert.Equal(t, int64(-1), attr.Value, "Attribute %s value should be -1", key)
		assert.Equal(t, int64(-1), attr.Worst, "Attribute %s worst should be -1", key)
		assert.Equal(t, int64(-1), attr.RawValue, "Attribute %s rawValue should be -1", key)
	}
}

func TestGetSmartAttributes_HasDescriptions(t *testing.T) {
	attrs := GetSmartAttributes()

	// Verify key attributes have descriptions
	tempAttr := attrs["temperature_celsius"]
	assert.NotEmpty(t, tempAttr.Description, "temperature_celsius should have a description")
	assert.Equal(t, "Celsius", tempAttr.Unit)

	powerAttr := attrs["power_on_hours"]
	assert.NotEmpty(t, powerAttr.Description, "power_on_hours should have a description")
	assert.Equal(t, "hours", powerAttr.Unit)

	percentAttr := attrs["percentage_used"]
	assert.NotEmpty(t, percentAttr.Description, "percentage_used should have a description")
	assert.Equal(t, "percent", percentAttr.Unit)
}

func TestGetSmartAttributes_NVMeSpecificAttributes(t *testing.T) {
	attrs := GetSmartAttributes()

	// Verify NVMe-specific attributes exist
	nvmeKeys := []string{
		"critical_warning",
		"nvme_error_log_entries",
		"nvme_media_errors",
		"available_spare",
		"available_spare_threshold",
		"media_and_data_integrity_errors",
		"unsafe_shutdowns",
		"host_read_commands",
		"host_write_commands",
		"controller_busy_time",
	}

	for _, key := range nvmeKeys {
		_, exists := attrs[key]
		assert.True(t, exists, "NVMe attribute %q should exist", key)
	}
}

func TestCleanupSmartAttributes_RemovesUnsetAttributes(t *testing.T) {
	attrs := GetSmartAttributes()

	// All attributes have default -1 values, so all should be removed
	CleanupSmartAttributes(attrs)

	assert.Empty(t, attrs, "All unset attributes should be removed")
}

func TestCleanupSmartAttributes_KeepsSetAttributes(t *testing.T) {
	attrs := GetSmartAttributes()

	// Set some values for temperature_celsius
	if tempAttr, exists := attrs["temperature_celsius"]; exists {
		tempAttr.Value = 42
		tempAttr.RawValue = 42
		attrs["temperature_celsius"] = tempAttr
	}

	// Set some values for power_on_hours
	if powerAttr, exists := attrs["power_on_hours"]; exists {
		powerAttr.Value = 100
		powerAttr.RawValue = 8760
		attrs["power_on_hours"] = powerAttr
	}

	CleanupSmartAttributes(attrs)

	// Only the set attributes should remain
	assert.Len(t, attrs, 2)
	assert.Contains(t, attrs, "temperature_celsius")
	assert.Contains(t, attrs, "power_on_hours")
}

func TestCleanupSmartAttributes_PartiallySetAttribute(t *testing.T) {
	attrs := map[string]SmartAttribute{
		"test_attr_all_set": {
			Description: "Test attribute",
			Unit:        "count",
			Threshold:   10,
			Value:       50,
			Worst:       45,
			RawValue:    50,
		},
		"test_attr_partial": {
			Description: "Partial attribute",
			Unit:        "count",
			Threshold:   -1,
			Value:       50, // Only value is set
			Worst:       -1,
			RawValue:    -1,
		},
		"test_attr_unset": {
			Description: "Unset attribute",
			Unit:        "count",
			Threshold:   -1,
			Value:       -1,
			Worst:       -1,
			RawValue:    -1,
		},
	}

	CleanupSmartAttributes(attrs)

	// All set and partially set attributes should remain
	assert.Len(t, attrs, 2)
	assert.Contains(t, attrs, "test_attr_all_set")
	assert.Contains(t, attrs, "test_attr_partial")
	assert.NotContains(t, attrs, "test_attr_unset")
}

func TestCleanupSmartAttributes_EmptyMap(t *testing.T) {
	attrs := map[string]SmartAttribute{}

	// Should not panic on empty map
	assert.NotPanics(t, func() {
		CleanupSmartAttributes(attrs)
	})

	assert.Empty(t, attrs)
}

func TestAliasMap_ContainsExpectedMappings(t *testing.T) {
	// Verify the alias map contains expected mappings
	assert.Equal(t, "temperature_celsius", aliasMap["current_drive_temperature"])
	assert.Equal(t, "unsafe_shutdowns", aliasMap["unsafe_shutdown_count"])
}

func TestSmartAttribute_StructFields(t *testing.T) {
	attr := SmartAttribute{
		Description: "Test Description",
		Unit:        "count",
		Threshold:   10,
		Value:       50,
		Worst:       45,
		RawValue:    500,
	}

	assert.Equal(t, "Test Description", attr.Description)
	assert.Equal(t, "count", attr.Unit)
	assert.Equal(t, int64(10), attr.Threshold)
	assert.Equal(t, int64(50), attr.Value)
	assert.Equal(t, int64(45), attr.Worst)
	assert.Equal(t, int64(500), attr.RawValue)
}

func TestDeviceInfo_StructFields(t *testing.T) {
	info := DeviceInfo{
		ModelFamily:       "Test Family",
		DeviceModel:       "Test Model",
		SerialNumber:      "ABC123",
		FirmwareVersion:   "1.0.0",
		Vendor:            "TestVendor",
		VendorID:          "0x1234",
		SubsystemVendorID: "0x5678",
		Product:           "TestProduct",
		LunID:             "lun123",
		Capacity:          1000.0,
		DWPD:              3.0,
		RPM:               7200,
		FormFactor:        "lff",
		Media:             "hdd",
		HealthStatus:      true,
	}

	assert.Equal(t, "Test Family", info.ModelFamily)
	assert.Equal(t, "Test Model", info.DeviceModel)
	assert.Equal(t, "ABC123", info.SerialNumber)
	assert.Equal(t, "1.0.0", info.FirmwareVersion)
	assert.Equal(t, "TestVendor", info.Vendor)
	assert.Equal(t, "0x1234", info.VendorID)
	assert.Equal(t, "0x5678", info.SubsystemVendorID)
	assert.Equal(t, "TestProduct", info.Product)
	assert.Equal(t, "lun123", info.LunID)
	assert.Equal(t, 1000.0, info.Capacity)
	assert.Equal(t, 3.0, info.DWPD)
	assert.Equal(t, int64(7200), info.RPM)
	assert.Equal(t, "lff", info.FormFactor)
	assert.Equal(t, "hdd", info.Media)
	assert.True(t, info.HealthStatus)
}

func TestNormalizedSmartData_StructFields(t *testing.T) {
	temp := int64(35)
	reallocated := int64(0)
	pending := int64(0)
	powerOn := int64(8760)
	ssdLife := int64(5)

	data := NormalizedSmartData{
		NodeName:           "node1",
		InstanceID:         "instance1",
		Device:             "/dev/sda",
		DeviceInfo:         &DeviceInfo{Vendor: "Intel"},
		CapacityGB:         480.0,
		TemperatureCelsius: &temp,
		ReallocatedSectors: &reallocated,
		PendingSectors:     &pending,
		PowerOnHours:       &powerOn,
		SSDLifeUsed:        &ssdLife,
		ErrorCounts:        map[string]int64{"UDMA_CRC_Error_Count": 0},
		Attributes:         map[string]SmartAttribute{},
		OSDID:              "osd.0",
	}

	assert.Equal(t, "node1", data.NodeName)
	assert.Equal(t, "instance1", data.InstanceID)
	assert.Equal(t, "/dev/sda", data.Device)
	assert.Equal(t, "Intel", data.DeviceInfo.Vendor)
	assert.Equal(t, 480.0, data.CapacityGB)
	assert.Equal(t, int64(35), *data.TemperatureCelsius)
	assert.Equal(t, int64(0), *data.ReallocatedSectors)
	assert.Equal(t, int64(0), *data.PendingSectors)
	assert.Equal(t, int64(8760), *data.PowerOnHours)
	assert.Equal(t, int64(5), *data.SSDLifeUsed)
	assert.Equal(t, "osd.0", data.OSDID)
}

func TestNatsEvent_StructFields(t *testing.T) {
	event := NatsEvent{
		NodeName:   "node1",
		InstanceID: "instance1",
		Device:     "/dev/sda",
		EventType:  "health_alert",
		Severity:   "critical",
		Message:    "Disk health degraded",
		Details: map[string]string{
			"reallocated_sectors": "100",
			"temperature":         "65",
		},
	}

	assert.Equal(t, "node1", event.NodeName)
	assert.Equal(t, "instance1", event.InstanceID)
	assert.Equal(t, "/dev/sda", event.Device)
	assert.Equal(t, "health_alert", event.EventType)
	assert.Equal(t, "critical", event.Severity)
	assert.Equal(t, "Disk health degraded", event.Message)
	assert.Equal(t, "100", event.Details["reallocated_sectors"])
	assert.Equal(t, "65", event.Details["temperature"])
}
