// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGigabytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{name: "whole number", input: "100", expected: 100},
		{name: "decimal", input: "1000.5", expected: 1000},
		{name: "zero", input: "0", expected: 0},
		{name: "empty string", input: "", expected: 0},
		{name: "invalid string", input: "abc", expected: 0},
		{name: "large number", input: "50000.99", expected: 50000},
		{name: "negative", input: "-10", expected: -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGigabytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculatePercentageUsed(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{name: "full life remaining", input: 100, expected: 0},
		{name: "no life remaining", input: 0, expected: 100},
		{name: "half life", input: 50, expected: 50},
		{name: "mostly used", input: 10, expected: 90},
		{name: "slightly used", input: 95, expected: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePercentageUsed(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateAttributeFromValue(t *testing.T) {
	t.Run("updates existing attribute", func(t *testing.T) {
		smartAttrs := GetSmartAttributes()
		updateAttributeFromValue(smartAttrs, "power_on_hours", 5000, 5000, -1, -1, "hours")

		attr, found := smartAttrs["power_on_hours"]
		assert.True(t, found)
		assert.Equal(t, int64(5000), attr.Value)
		assert.Equal(t, int64(5000), attr.RawValue)
		assert.Equal(t, "hours", attr.Unit)
	})

	t.Run("updates threshold when provided", func(t *testing.T) {
		smartAttrs := GetSmartAttributes()
		updateAttributeFromValue(smartAttrs, "temperature_celsius", 35, 35, 70, -1, "Celsius")

		attr := smartAttrs["temperature_celsius"]
		assert.Equal(t, int64(35), attr.Value)
		assert.Equal(t, int64(70), attr.Threshold)
	})

	t.Run("updates worst when provided", func(t *testing.T) {
		smartAttrs := GetSmartAttributes()
		updateAttributeFromValue(smartAttrs, "temperature_celsius", 35, 35, -1, 50, "Celsius")

		attr := smartAttrs["temperature_celsius"]
		assert.Equal(t, int64(50), attr.Worst)
	})

	t.Run("ignores threshold -1", func(t *testing.T) {
		smartAttrs := GetSmartAttributes()
		originalThreshold := smartAttrs["temperature_celsius"].Threshold
		updateAttributeFromValue(smartAttrs, "temperature_celsius", 35, 35, -1, -1, "Celsius")

		assert.Equal(t, originalThreshold, smartAttrs["temperature_celsius"].Threshold)
	})

	t.Run("ignores unknown attribute", func(t *testing.T) {
		smartAttrs := GetSmartAttributes()
		origLen := len(smartAttrs)
		updateAttributeFromValue(smartAttrs, "nonexistent_attr", 100, 100, -1, -1, "")
		assert.Equal(t, origLen, len(smartAttrs))
	})
}

func TestProcessAndUpdateSmartAttributes_ATA(t *testing.T) {
	smartAttrs := GetSmartAttributes()

	output := &SmartCtlOutput{
		ATASMARTAttributes: &SmartCtlATASMARTAttributes{
			Revision: 16,
			Table: []SmartCtlATASMARTEntry{
				{
					ID:     5,
					Name:   "Reallocated_Sector_Ct",
					Value:  100,
					Worst:  100,
					Thresh: 10,
					Raw:    SmartCtlATASMARTRaw{Value: 3, String: "3"},
				},
				{
					ID:     9,
					Name:   "Power_On_Hours",
					Value:  95,
					Worst:  95,
					Thresh: 0,
					Raw:    SmartCtlATASMARTRaw{Value: 10000, String: "10000"},
				},
			},
		},
		PowerOnTime: SmartCtlPowerOnTime{Hours: 10000},
		Temperature: SmartCtlTemperature{Current: 35},
	}

	ProcessAndUpdateSmartAttributes(smartAttrs, output)

	// ATA attributes should be updated
	if attr, ok := smartAttrs["reallocated_sector_ct"]; ok {
		assert.Equal(t, int64(100), attr.Value)
		assert.Equal(t, int64(3), attr.RawValue)
	}
}

func TestProcessAndUpdateSCSISmartAttributes(t *testing.T) {
	smartAttrs := GetSmartAttributes()

	output := &SmartCtlOutput{
		PowerOnTime: SmartCtlPowerOnTime{Hours: 5000},
		Temperature: SmartCtlTemperature{Current: 35},
		SCSIGrownDefectList: 10,
		SCSIStartStopCycleCounter: &SmartCtlSCSIStartStopCycle{
			AccumulatedStartStopCycles: 200,
		},
		SCSIErrorCounterLog: &SmartCtlSCSIErrorCounterLog{
			Read: SmartCtlSCSIErrorDetails{
				TotalErrorsCorrected:   100,
				TotalUncorrectedErrors: 0,
				GigabytesProcessed:     "5000.5",
			},
			Write: SmartCtlSCSIErrorDetails{
				TotalErrorsCorrected:   50,
				TotalUncorrectedErrors: 0,
				GigabytesProcessed:     "2500.3",
			},
			Verify: SmartCtlSCSIErrorDetails{
				TotalErrorsCorrected:   5,
				TotalUncorrectedErrors: 0,
			},
		},
	}

	ProcessAndUpdateSCSISmartAttributes(smartAttrs, output)

	// Power on hours
	if attr, ok := smartAttrs["power_on_hours"]; ok {
		assert.Equal(t, int64(5000), attr.Value)
	}

	// Temperature
	if attr, ok := smartAttrs["temperature_celsius"]; ok {
		assert.Equal(t, int64(35), attr.Value)
	}

	// Grown defects
	if attr, ok := smartAttrs["grown_defects_count"]; ok {
		assert.Equal(t, int64(10), attr.Value)
	}

	// Power cycle count
	if attr, ok := smartAttrs["power_cycle_count"]; ok {
		assert.Equal(t, int64(200), attr.Value)
	}
}

func TestProcessAndUpdateSCSISmartAttributes_ZeroTemperature(t *testing.T) {
	smartAttrs := GetSmartAttributes()
	originalValue := smartAttrs["temperature_celsius"].Value
	output := &SmartCtlOutput{
		Temperature: SmartCtlTemperature{Current: 0},
	}

	// Should not update temperature for 0°C
	ProcessAndUpdateSCSISmartAttributes(smartAttrs, output)

	if attr, ok := smartAttrs["temperature_celsius"]; ok {
		assert.Equal(t, originalValue, attr.Value) // not updated from original
	}
}

func TestProcessAndUpdateNVMeSmartAttributes(t *testing.T) {
	smartAttrs := GetSmartAttributes()

	output := &SmartCtlOutput{
		NVMeSmartHealthInfoLog: &SmartCtlNVMeSmartHealthInfoLog{
			PowerOnHours:            10000,
			Temperature:             40,
			PowerCycles:             500,
			UnsafeShutdowns:         5,
			HostReads:               2000000,
			HostWrites:              1000000,
			ControllerBusyTime:      300,
			NumErrLogEntries:        2,
			PercentageUsed:          10,
			AvailableSpare:          90,
			AvailableSpareThreshold: 10,
			MediaErrors:             0,
			CriticalWarning:         0,
		},
	}

	ProcessAndUpdateNVMeSmartAttributes(smartAttrs, output)

	if attr, ok := smartAttrs["power_on_hours"]; ok {
		assert.Equal(t, int64(10000), attr.Value)
	}
	if attr, ok := smartAttrs["temperature_celsius"]; ok {
		assert.Equal(t, int64(40), attr.Value)
	}
	if attr, ok := smartAttrs["power_cycle_count"]; ok {
		assert.Equal(t, int64(500), attr.Value)
	}
	if attr, ok := smartAttrs["percentage_used"]; ok {
		assert.Equal(t, int64(10), attr.Value)
	}
	if attr, ok := smartAttrs["available_spare"]; ok {
		assert.Equal(t, int64(90), attr.Value)
	}
}

func TestProcessAndUpdateNVMeSmartAttributes_NilHealthLog(t *testing.T) {
	smartAttrs := GetSmartAttributes()
	output := &SmartCtlOutput{
		NVMeSmartHealthInfoLog: nil,
	}

	// Should not panic
	ProcessAndUpdateNVMeSmartAttributes(smartAttrs, output)
}

func TestFindSmartAttributeByID(t *testing.T) {
	attributes := []SmartCtlATASMARTEntry{
		{ID: 1, Name: "Raw_Read_Error_Rate", Value: 100},
		{ID: 5, Name: "Reallocated_Sector_Ct", Value: 98},
		{ID: 9, Name: "Power_On_Hours", Value: 95},
		{ID: 197, Name: "Current_Pending_Sector", Value: 100},
	}

	t.Run("finds existing attribute", func(t *testing.T) {
		entry := findSmartAttributeByID(attributes, 5)
		assert.NotNil(t, entry)
		assert.Equal(t, "Reallocated_Sector_Ct", entry.Name)
		assert.Equal(t, int64(98), entry.Value)
	})

	t.Run("finds first attribute", func(t *testing.T) {
		entry := findSmartAttributeByID(attributes, 1)
		assert.NotNil(t, entry)
		assert.Equal(t, "Raw_Read_Error_Rate", entry.Name)
	})

	t.Run("finds last attribute", func(t *testing.T) {
		entry := findSmartAttributeByID(attributes, 197)
		assert.NotNil(t, entry)
		assert.Equal(t, "Current_Pending_Sector", entry.Name)
	})

	t.Run("returns nil for missing attribute", func(t *testing.T) {
		entry := findSmartAttributeByID(attributes, 999)
		assert.Nil(t, entry)
	})

	t.Run("empty slice", func(t *testing.T) {
		entry := findSmartAttributeByID([]SmartCtlATASMARTEntry{}, 1)
		assert.Nil(t, entry)
	})
}

func TestProcessAndUpdateATASmartAttributes_PercentageAttributes(t *testing.T) {
	smartAttrs := GetSmartAttributes()

	output := &SmartCtlOutput{
		ATASMARTAttributes: &SmartCtlATASMARTAttributes{
			Table: []SmartCtlATASMARTEntry{
				{
					ID:    233,
					Name:  "Media_Wearout_Indicator",
					Value: 90, // 90% life remaining
					Raw:   SmartCtlATASMARTRaw{Value: 90},
				},
			},
		},
	}

	ProcessAndUpdateATASmartAttributes(smartAttrs, output)

	// media_wearout_indicator should be calculated as 100 - value = 10% used
	if attr, ok := smartAttrs["media_wearout_indicator"]; ok {
		assert.Equal(t, int64(10), attr.Value) // 100 - 90 = 10
	}
}

func TestNormalizeSmartData_ATA(t *testing.T) {
	smartData := &SmartCtlOutput{
		Device: SmartCtlDevice{
			Name:     "/dev/sda",
			Protocol: "ATA",
			Type:     "sat",
		},
		ModelName:       "WDC WD10EZEX-00BN5A0",
		SerialNumber:    "WD-WMC3T0123456",
		FirmwareVersion: "01.01A01",
		Temperature:     SmartCtlTemperature{Current: 35},
		PowerOnTime:     SmartCtlPowerOnTime{Hours: 10000},
		UserCapacity: &SmartCtlUserCapacity{
			Bytes: 1000204886016,
		},
		ATASMARTAttributes: &SmartCtlATASMARTAttributes{
			Table: []SmartCtlATASMARTEntry{
				{ID: 5, Name: "Reallocated_Sector_Ct", Value: 100, Raw: SmartCtlATASMARTRaw{Value: 3}},
				{ID: 197, Name: "Current_Pending_Sector", Value: 100, Raw: SmartCtlATASMARTRaw{Value: 1}},
				{ID: 199, Name: "UDMA_CRC_Error_Count", Value: 100, Raw: SmartCtlATASMARTRaw{Value: 5}},
			},
		},
		SmartStatus: SmartCtlSmartStatus{Passed: true},
	}

	deviceInfo := &DeviceInfo{Capacity: -1}
	FillDeviceInfoFromSmartData(deviceInfo, smartData)
	smartAttrs := GetSmartAttributes()

	result := normalizeSmartData(smartData, deviceInfo, smartAttrs, "node1", "instance1", "")

	assert.Equal(t, "node1", result.NodeName)
	assert.Equal(t, "instance1", result.InstanceID)
	assert.Equal(t, "/dev/sda", result.Device)
	assert.NotNil(t, result.TemperatureCelsius)
	assert.Equal(t, int64(35), *result.TemperatureCelsius)
	assert.NotNil(t, result.ReallocatedSectors)
	assert.Equal(t, int64(3), *result.ReallocatedSectors)
	assert.NotNil(t, result.PendingSectors)
	assert.Equal(t, int64(1), *result.PendingSectors)
	assert.NotNil(t, result.PowerOnHours)
	assert.Equal(t, int64(10000), *result.PowerOnHours)
	assert.Equal(t, int64(5), result.ErrorCounts["UDMA_CRC_Error_Count"])
	assert.True(t, result.CapacityGB > 0)
}

func TestNormalizeSmartData_NVMe(t *testing.T) {
	smartData := &SmartCtlOutput{
		Device: SmartCtlDevice{
			Name:     "/dev/nvme0n1",
			Protocol: "NVMe",
			Type:     "nvme",
		},
		ModelName:       "Samsung SSD 980 PRO",
		SerialNumber:    "S5GXNF0N123456",
		FirmwareVersion: "5B2QGXA7",
		Temperature:     SmartCtlTemperature{Current: 40},
		NVMeSmartHealthInfoLog: &SmartCtlNVMeSmartHealthInfoLog{
			PercentageUsed: 5,
		},
		SmartStatus: SmartCtlSmartStatus{Passed: true},
	}

	deviceInfo := &DeviceInfo{Capacity: 1000.0}
	smartAttrs := GetSmartAttributes()

	result := normalizeSmartData(smartData, deviceInfo, smartAttrs, "node2", "instance2", "")

	assert.Equal(t, "node2", result.NodeName)
	assert.Equal(t, "/dev/nvme0n1", result.Device)
	assert.NotNil(t, result.TemperatureCelsius)
	assert.Equal(t, int64(40), *result.TemperatureCelsius)
	assert.NotNil(t, result.SSDLifeUsed)
	assert.Equal(t, int64(5), *result.SSDLifeUsed)
}

func TestNormalizeSmartData_SCSI(t *testing.T) {
	smartData := &SmartCtlOutput{
		Device: SmartCtlDevice{
			Name:     "/dev/sdb",
			Protocol: "SCSI",
			Type:     "scsi",
		},
		Temperature:     SmartCtlTemperature{Current: 30},
		PowerOnTime:     SmartCtlPowerOnTime{Hours: 20000},
		SCSIGrownDefectList: 5,
		SCSIStartStopCycleCounter: &SmartCtlSCSIStartStopCycle{
			AccumulatedStartStopCycles: 100,
		},
		SmartStatus: SmartCtlSmartStatus{Passed: true},
	}

	deviceInfo := &DeviceInfo{Capacity: 4000.0}
	smartAttrs := GetSmartAttributes()

	result := normalizeSmartData(smartData, deviceInfo, smartAttrs, "node3", "instance3", "")

	assert.Equal(t, "/dev/sdb", result.Device)
	assert.NotNil(t, result.TemperatureCelsius)
	assert.Equal(t, int64(30), *result.TemperatureCelsius)
	assert.NotNil(t, result.PowerOnHours)
	assert.Equal(t, int64(20000), *result.PowerOnHours)
	assert.NotNil(t, result.ReallocatedSectors)
	assert.Equal(t, int64(5), *result.ReallocatedSectors)
}

func TestNormalizeSmartData_ZeroTemperature(t *testing.T) {
	smartData := &SmartCtlOutput{
		Device: SmartCtlDevice{
			Name:     "/dev/sda",
			Protocol: "ATA",
		},
		Temperature: SmartCtlTemperature{Current: 0},
		SmartStatus: SmartCtlSmartStatus{Passed: true},
	}

	deviceInfo := &DeviceInfo{Capacity: 500.0}
	smartAttrs := GetSmartAttributes()

	result := normalizeSmartData(smartData, deviceInfo, smartAttrs, "node", "inst", "")

	assert.Nil(t, result.TemperatureCelsius) // 0 temp should result in nil
}

func TestCollectSmartDataFromFile_ValidJSON(t *testing.T) {
	// Create a temp file with valid smartctl JSON
	tmpFile, err := os.CreateTemp("", "smartctl-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	jsonContent := `{
		"json_format_version": [1, 0],
		"smartctl": {"version": [7, 4], "exit_status": 0},
		"device": {"name": "/dev/sda", "type": "sat", "protocol": "ATA"},
		"model_name": "Test SSD 1TB",
		"serial_number": "TEST123",
		"firmware_version": "1.0",
		"smart_status": {"passed": true},
		"smart_support": {"available": true, "enabled": true},
		"temperature": {"current": 35}
	}`

	_, err = tmpFile.WriteString(jsonContent)
	assert.NoError(t, err)
	tmpFile.Close()

	result, err := collectSmartDataFromFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test SSD 1TB", result.ModelName)
	assert.Equal(t, "TEST123", result.SerialNumber)
	assert.Equal(t, "/dev/sda", result.Device.Name)
	assert.True(t, result.SmartStatus.Passed)
	assert.Equal(t, int64(35), result.Temperature.Current)
}

func TestCollectSmartDataFromFile_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "smartctl-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("not valid json")
	assert.NoError(t, err)
	tmpFile.Close()

	_, err = collectSmartDataFromFile(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing JSON")
}

func TestCollectSmartDataFromFile_NonExistentFile(t *testing.T) {
	_, err := collectSmartDataFromFile("/nonexistent/path/to/file.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading file")
}

func TestCollectSmartDataFromFile_NVMeDevice(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "smartctl-nvme-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	jsonContent := `{
		"json_format_version": [1, 0],
		"smartctl": {"version": [7, 4], "exit_status": 0},
		"device": {"name": "/dev/nvme0n1", "type": "nvme", "protocol": "NVMe"},
		"model_name": "Samsung SSD 980 PRO",
		"serial_number": "S5GXNF0N123456",
		"firmware_version": "5B2QGXA7",
		"smart_status": {"passed": true, "nvme": {"value": 0}},
		"smart_support": {"available": true, "enabled": true},
		"temperature": {"current": 40},
		"nvme_smart_health_information_log": {
			"available_spare": 100,
			"available_spare_threshold": 10,
			"percentage_used": 5,
			"media_errors": 0,
			"num_err_log_entries": 0,
			"power_on_hours": 5000,
			"power_cycles": 200,
			"temperature": 40
		},
		"nvme_pci_vendor": {"id": 5197, "subsystem_id": 5197},
		"nvme_total_capacity": 1000204886016
	}`

	_, err = tmpFile.WriteString(jsonContent)
	assert.NoError(t, err)
	tmpFile.Close()

	result, err := collectSmartDataFromFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "NVMe", result.Device.Protocol)
	assert.NotNil(t, result.NVMeSmartHealthInfoLog)
	assert.Equal(t, int64(5), result.NVMeSmartHealthInfoLog.PercentageUsed)
	assert.Equal(t, int64(5000), result.NVMeSmartHealthInfoLog.PowerOnHours)
	assert.NotNil(t, result.NVMePCIVendor)
	assert.Equal(t, int64(5197), result.NVMePCIVendor.ID)
}
