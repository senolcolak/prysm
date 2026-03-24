// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSmartCtlScanOutput_StructFields(t *testing.T) {
	output := SmartCtlScanOutput{
		JSONFormatVersion: []int64{1, 0},
		Smartctl: SmartCtlDetails{
			Version:   []int64{7, 4},
			BuildInfo: "build info",
		},
		Devices: []SmartCtlDevice{
			{Name: "/dev/sda", Type: "sat", Protocol: "ATA"},
			{Name: "/dev/nvme0", Type: "nvme", Protocol: "NVMe"},
		},
	}

	assert.Equal(t, []int64{1, 0}, output.JSONFormatVersion)
	assert.Len(t, output.Devices, 2)
	assert.Equal(t, "/dev/sda", output.Devices[0].Name)
	assert.Equal(t, "nvme", output.Devices[1].Type)
}

func TestSmartCtlScanOutput_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"json_format_version": [1, 0],
		"smartctl": {
			"version": [7, 4],
			"build_info": "test build",
			"exit_status": 0
		},
		"devices": [
			{"name": "/dev/sda", "info_name": "/dev/sda [SAT]", "type": "sat", "protocol": "ATA"},
			{"name": "/dev/nvme0", "info_name": "/dev/nvme0", "type": "nvme", "protocol": "NVMe"}
		]
	}`

	var output SmartCtlScanOutput
	err := json.Unmarshal([]byte(jsonData), &output)
	assert.NoError(t, err)

	assert.Equal(t, []int64{1, 0}, output.JSONFormatVersion)
	assert.Len(t, output.Devices, 2)
	assert.Equal(t, "/dev/sda", output.Devices[0].Name)
	assert.Equal(t, "ATA", output.Devices[0].Protocol)
}

func TestSmartCtlDevice_StructFields(t *testing.T) {
	device := SmartCtlDevice{
		InfoName: "/dev/sda [SAT]",
		Name:     "/dev/sda",
		Protocol: "ATA",
		Type:     "sat",
	}

	assert.Equal(t, "/dev/sda [SAT]", device.InfoName)
	assert.Equal(t, "/dev/sda", device.Name)
	assert.Equal(t, "ATA", device.Protocol)
	assert.Equal(t, "sat", device.Type)
}

func TestSmartCtlDevice_JSONRoundTrip(t *testing.T) {
	original := SmartCtlDevice{
		InfoName: "/dev/nvme0n1",
		Name:     "/dev/nvme0n1",
		Protocol: "NVMe",
		Type:     "nvme",
	}

	jsonData, err := json.Marshal(original)
	assert.NoError(t, err)

	var restored SmartCtlDevice
	err = json.Unmarshal(jsonData, &restored)
	assert.NoError(t, err)

	assert.Equal(t, original, restored)
}

func TestSmartCtlATAVersion_StructFields(t *testing.T) {
	ataVersion := SmartCtlATAVersion{
		MajorValue: 10,
		MinorValue: 0,
		String:     "ACS-3 T13/2161-D revision 5",
	}

	assert.Equal(t, int64(10), ataVersion.MajorValue)
	assert.Equal(t, int64(0), ataVersion.MinorValue)
	assert.Contains(t, ataVersion.String, "ACS-3")
}

func TestSmartCtlATASMARTAttributes_StructFields(t *testing.T) {
	attrs := SmartCtlATASMARTAttributes{
		Revision: 16,
		Table: []SmartCtlATASMARTEntry{
			{
				ID:     1,
				Name:   "Raw_Read_Error_Rate",
				Value:  100,
				Worst:  100,
				Thresh: 50,
				Flags: SmartCtlATASMARTFlags{
					Value:         0x000f,
					String:        "POSR-",
					Prefailure:    true,
					UpdatedOnline: true,
				},
				Raw: SmartCtlATASMARTRaw{
					Value:  0,
					String: "0",
				},
			},
		},
	}

	assert.Equal(t, int64(16), attrs.Revision)
	assert.Len(t, attrs.Table, 1)
	assert.Equal(t, int64(1), attrs.Table[0].ID)
	assert.Equal(t, "Raw_Read_Error_Rate", attrs.Table[0].Name)
	assert.True(t, attrs.Table[0].Flags.Prefailure)
}

func TestSmartCtlATASMARTEntry_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"id": 5,
		"name": "Reallocated_Sector_Ct",
		"value": 100,
		"worst": 100,
		"thresh": 10,
		"when_failed": "",
		"flags": {
			"value": 51,
			"string": "PO--CK",
			"prefailure": true,
			"updated_online": false,
			"performance": false,
			"error_rate": false,
			"event_count": true,
			"auto_keep": true
		},
		"raw": {
			"value": 0,
			"string": "0"
		}
	}`

	var entry SmartCtlATASMARTEntry
	err := json.Unmarshal([]byte(jsonData), &entry)
	assert.NoError(t, err)

	assert.Equal(t, int64(5), entry.ID)
	assert.Equal(t, "Reallocated_Sector_Ct", entry.Name)
	assert.Equal(t, int64(100), entry.Value)
	assert.Equal(t, int64(10), entry.Thresh)
	assert.True(t, entry.Flags.Prefailure)
	assert.True(t, entry.Flags.EventCount)
	assert.True(t, entry.Flags.AutoKeep)
	assert.Equal(t, int64(0), entry.Raw.Value)
}

func TestSmartCtlATASMARTErrorLog_StructFields(t *testing.T) {
	errorLog := SmartCtlATASMARTErrorLog{
		Summary: SmartCtlATASMARTErrorLogSummary{
			Count:       5,
			Revision:    1,
			LoggedCount: 5,
			Table: []SmartCtlATASMARTErrorLogEntry{
				{
					ErrorNumber:      1,
					LifetimeHours:    1000,
					ErrorDescription: "UNC error at LBA",
					CompletionRegisters: SmartCtlATASMARTCompletionRegisters{
						Count:  1,
						Device: 0,
						Error:  64,
						LBA:    123456,
						Status: 81,
					},
				},
			},
		},
	}

	assert.Equal(t, int64(5), errorLog.Summary.Count)
	assert.Len(t, errorLog.Summary.Table, 1)
	assert.Equal(t, int64(123456), errorLog.Summary.Table[0].CompletionRegisters.LBA)
}

func TestSmartCtlATASMARTPreviousCommand_StructFields(t *testing.T) {
	cmd := SmartCtlATASMARTPreviousCommand{
		CommandName:         "READ DMA EXT",
		PowerupMilliseconds: 50000,
		Registers: SmartCtlATASMARTRegisters{
			Command:       37,
			Count:         8,
			Device:        64,
			DeviceControl: 0,
			Features:      0,
			LBA:           123456,
		},
	}

	assert.Equal(t, "READ DMA EXT", cmd.CommandName)
	assert.Equal(t, int64(50000), cmd.PowerupMilliseconds)
	assert.Equal(t, int64(37), cmd.Registers.Command)
}

func TestSmartCtlTrimSupport_StructFields(t *testing.T) {
	trim := SmartCtlTrimSupport{Supported: true}
	assert.True(t, trim.Supported)

	trim2 := SmartCtlTrimSupport{Supported: false}
	assert.False(t, trim2.Supported)
}

func TestSmartCtlDeviceType_StructFields(t *testing.T) {
	dt := SmartCtlDeviceType{
		Name:            "disk",
		SCSITerminology: "Direct access",
		SCSIValue:       0,
	}

	assert.Equal(t, "disk", dt.Name)
	assert.Equal(t, "Direct access", dt.SCSITerminology)
}

func TestSmartCtlFormFactor_StructFields(t *testing.T) {
	ff := SmartCtlFormFactor{
		Name:      "2.5 inches",
		SCSIValue: 3,
		ATAValue:  2,
	}

	assert.Equal(t, "2.5 inches", ff.Name)
	assert.Equal(t, int64(3), ff.SCSIValue)
	assert.Equal(t, int64(2), ff.ATAValue)
}

func TestSmartCtlInterfaceSpeed_StructFields(t *testing.T) {
	speed := SmartCtlInterfaceSpeed{
		Max: SmartCtlSpeedInfo{
			BitsPerUnit:    8,
			SATAValue:      3,
			String:         "6.0 Gb/s",
			UnitsPerSecond: 600000000,
		},
		Current: SmartCtlSpeedInfo{
			BitsPerUnit:    8,
			SATAValue:      3,
			String:         "6.0 Gb/s",
			UnitsPerSecond: 600000000,
		},
	}

	assert.Equal(t, "6.0 Gb/s", speed.Max.String)
	assert.Equal(t, int64(600000000), speed.Current.UnitsPerSecond)
}

func TestSmartCtlLocalTime_StructFields(t *testing.T) {
	lt := SmartCtlLocalTime{
		Asctime: "Mon Jan 15 12:30:45 2024",
		TimeT:   1705322445,
	}

	assert.Contains(t, lt.Asctime, "Jan")
	assert.Equal(t, int64(1705322445), lt.TimeT)
}

func TestSmartCtlPowerOnTime_StructFields(t *testing.T) {
	pot := SmartCtlPowerOnTime{
		Hours:   10000,
		Minutes: 30,
	}

	assert.Equal(t, int64(10000), pot.Hours)
	assert.Equal(t, int64(30), pot.Minutes)
}

func TestSmartCtlSATAVersion_StructFields(t *testing.T) {
	sv := SmartCtlSATAVersion{
		String: "SATA 3.2",
		Value:  50,
	}

	assert.Equal(t, "SATA 3.2", sv.String)
	assert.Equal(t, int64(50), sv.Value)
}

func TestSmartCtlSCSIErrorCounterLog_StructFields(t *testing.T) {
	ecl := SmartCtlSCSIErrorCounterLog{
		Read: SmartCtlSCSIErrorDetails{
			CorrectionAlgorithmInvocations: 100,
			ErrorsCorrectedByECCDelayed:    5,
			ErrorsCorrectedByECCFast:       10,
			ErrorsCorrectedByReReads:       2,
			GigabytesProcessed:             "1000.5",
			TotalErrorsCorrected:           17,
			TotalUncorrectedErrors:         0,
		},
		Verify: SmartCtlSCSIErrorDetails{
			TotalUncorrectedErrors: 0,
		},
		Write: SmartCtlSCSIErrorDetails{
			TotalErrorsCorrected: 3,
		},
	}

	assert.Equal(t, int64(100), ecl.Read.CorrectionAlgorithmInvocations)
	assert.Equal(t, "1000.5", ecl.Read.GigabytesProcessed)
	assert.Equal(t, int64(17), ecl.Read.TotalErrorsCorrected)
	assert.Equal(t, int64(0), ecl.Write.TotalUncorrectedErrors)
}

func TestSmartCtlSCSIStartStopCycle_StructFields(t *testing.T) {
	ssc := SmartCtlSCSIStartStopCycle{
		AccumulatedLoadUnloadCycles:                500,
		AccumulatedStartStopCycles:                 100,
		SpecifiedCycleCountOverDeviceLifetime:      50000,
		SpecifiedLoadUnloadCountOverDeviceLifetime: 600000,
		WeekOfManufacture:                          "42",
		YearOfManufacture:                          "2023",
	}

	assert.Equal(t, int64(500), ssc.AccumulatedLoadUnloadCycles)
	assert.Equal(t, int64(100), ssc.AccumulatedStartStopCycles)
	assert.Equal(t, "2023", ssc.YearOfManufacture)
}

func TestSmartCtlSCSITransportProtocol_StructFields(t *testing.T) {
	tp := SmartCtlSCSITransportProtocol{
		Name:  "SAS (SPL-4)",
		Value: 6,
	}

	assert.Equal(t, "SAS (SPL-4)", tp.Name)
	assert.Equal(t, int64(6), tp.Value)
}

func TestSmartCtlNVMePCIVendor_StructFields(t *testing.T) {
	vendor := SmartCtlNVMePCIVendor{
		ID:          0x144D, // Samsung
		SubsystemID: 0x144D,
	}

	assert.Equal(t, int64(0x144D), vendor.ID)
	assert.Equal(t, int64(0x144D), vendor.SubsystemID)
}

func TestSmartCtlNVMeSmartHealthInfoLog_StructFields(t *testing.T) {
	health := SmartCtlNVMeSmartHealthInfoLog{
		AvailableSpare:          100,
		AvailableSpareThreshold: 10,
		ControllerBusyTime:      500,
		CriticalCompTime:        0,
		CriticalWarning:         0,
		DataUnitsRead:           1000000,
		DataUnitsWritten:        500000,
		HostReads:               2000000,
		HostWrites:              1000000,
		MediaErrors:             0,
		NumErrLogEntries:        0,
		PercentageUsed:          5,
		PowerCycles:             100,
		PowerOnHours:            5000,
		Temperature:             35,
		TemperatureSensors:      []int64{35, 38, 40},
		UnsafeShutdowns:         5,
		WarningTempTime:         0,
	}

	assert.Equal(t, int64(100), health.AvailableSpare)
	assert.Equal(t, int64(5), health.PercentageUsed)
	assert.Equal(t, int64(35), health.Temperature)
	assert.Len(t, health.TemperatureSensors, 3)
	assert.Equal(t, int64(0), health.MediaErrors)
}

func TestSmartCtlNVMeSmartHealthInfoLog_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"available_spare": 90,
		"available_spare_threshold": 10,
		"controller_busy_time": 1000,
		"critical_comp_time": 0,
		"critical_warning": 0,
		"data_units_read": 5000000,
		"data_units_written": 2500000,
		"host_reads": 10000000,
		"host_writes": 5000000,
		"media_errors": 2,
		"num_err_log_entries": 5,
		"percentage_used": 15,
		"power_cycles": 200,
		"power_on_hours": 10000,
		"temperature": 45,
		"temperature_sensors": [45, 48, 50, 52],
		"unsafe_shutdowns": 10,
		"warning_temp_time": 100
	}`

	var health SmartCtlNVMeSmartHealthInfoLog
	err := json.Unmarshal([]byte(jsonData), &health)
	assert.NoError(t, err)

	assert.Equal(t, int64(90), health.AvailableSpare)
	assert.Equal(t, int64(2), health.MediaErrors)
	assert.Equal(t, int64(15), health.PercentageUsed)
	assert.Len(t, health.TemperatureSensors, 4)
	assert.Equal(t, int64(100), health.WarningTempTime)
}

func TestSmartCtlNVMENamespace_StructFields(t *testing.T) {
	ns := SmartCtlNVMENamespace{
		ID: 1,
		Size: SmartCtlNVMeCapacity{
			Blocks: 1953525168,
			Bytes:  1000204886016,
		},
		Utilization: SmartCtlNVMeCapacity{
			Blocks: 1953525168,
			Bytes:  1000204886016,
		},
		Capacity: SmartCtlNVMeCapacity{
			Blocks: 1953525168,
			Bytes:  1000204886016,
		},
		FormattedLBASize: 512,
		EUI64: &SmartCtlNVMeEUI64{
			ExtID: 0,
			OUI:   0x002538,
		},
	}

	assert.Equal(t, int64(1), ns.ID)
	assert.Equal(t, int64(1000204886016), ns.Size.Bytes)
	assert.Equal(t, int64(512), ns.FormattedLBASize)
	assert.NotNil(t, ns.EUI64)
	assert.Equal(t, int64(0x002538), ns.EUI64.OUI)
}

func TestSmartCtlNVMeVersion_StructFields(t *testing.T) {
	version := SmartCtlNVMeVersion{
		String: "1.4",
		Value:  66304,
	}

	assert.Equal(t, "1.4", version.String)
	assert.Equal(t, int64(66304), version.Value)
}

func TestSmartCtlSmartStatus_StructFields(t *testing.T) {
	// Test passed status
	status := SmartCtlSmartStatus{
		Passed: true,
		NVMe: &SmartCtlNVMeStatus{
			Value: 0,
		},
	}

	assert.True(t, status.Passed)
	assert.NotNil(t, status.NVMe)
	assert.Equal(t, int64(0), status.NVMe.Value)

	// Test failed status
	failedStatus := SmartCtlSmartStatus{
		Passed: false,
	}
	assert.False(t, failedStatus.Passed)
	assert.Nil(t, failedStatus.NVMe)
}

func TestSmartCtlSmartSupport_StructFields(t *testing.T) {
	support := SmartCtlSmartSupport{
		Available: true,
		Enabled:   true,
	}

	assert.True(t, support.Available)
	assert.True(t, support.Enabled)

	// Test disabled
	disabled := SmartCtlSmartSupport{
		Available: true,
		Enabled:   false,
	}
	assert.True(t, disabled.Available)
	assert.False(t, disabled.Enabled)
}

func TestSmartCtlDetails_StructFields(t *testing.T) {
	details := SmartCtlDetails{
		Argv:         []string{"smartctl", "-a", "-j", "/dev/sda"},
		BuildInfo:    "(build info)",
		ExitStatus:   0,
		PlatformInfo: "x86_64-linux-gnu",
		SvnRevision:  "5338",
		Version:      []int64{7, 4},
		DriveDatabaseVersion: StringInfo{
			String: "7.4-2024.01.01",
		},
	}

	assert.Len(t, details.Argv, 4)
	assert.Equal(t, "smartctl", details.Argv[0])
	assert.Equal(t, int64(0), details.ExitStatus)
	assert.Equal(t, []int64{7, 4}, details.Version)
}

func TestSmartCtlTemperature_StructFields(t *testing.T) {
	temp := SmartCtlTemperature{
		Current:   35,
		DriveTrip: 70,
	}

	assert.Equal(t, int64(35), temp.Current)
	assert.Equal(t, int64(70), temp.DriveTrip)
}

func TestSmartCtlTemperatureWarning_StructFields(t *testing.T) {
	warning := SmartCtlTemperatureWarning{Enabled: true}
	assert.True(t, warning.Enabled)

	disabled := SmartCtlTemperatureWarning{Enabled: false}
	assert.False(t, disabled.Enabled)
}

func TestSmartCtlUserCapacity_StructFields(t *testing.T) {
	capacity := SmartCtlUserCapacity{
		Blocks: 1953525168,
		Bytes:  1000204886016,
	}

	assert.Equal(t, int64(1953525168), capacity.Blocks)
	assert.Equal(t, int64(1000204886016), capacity.Bytes)
}

func TestSmartCtlWWN_StructFields(t *testing.T) {
	wwn := SmartCtlWWN{
		ID:  0x5002538,
		NAA: 5,
		OUI: 0x002538,
	}

	assert.Equal(t, int64(0x5002538), wwn.ID)
	assert.Equal(t, int64(5), wwn.NAA)
	assert.Equal(t, int64(0x002538), wwn.OUI)
}

func TestSmartCtlOutput_MinimalFields(t *testing.T) {
	// Test with minimal required fields
	output := SmartCtlOutput{
		ModelName:       "Samsung SSD 980 PRO 1TB",
		SerialNumber:    "S5GXNF0N123456",
		FirmwareVersion: "5B2QGXA7",
		Device: SmartCtlDevice{
			Name:     "/dev/nvme0n1",
			Protocol: "NVMe",
			Type:     "nvme",
		},
		SmartStatus: SmartCtlSmartStatus{
			Passed: true,
		},
		SmartSupport: SmartCtlSmartSupport{
			Available: true,
			Enabled:   true,
		},
	}

	assert.Equal(t, "Samsung SSD 980 PRO 1TB", output.ModelName)
	assert.Equal(t, "S5GXNF0N123456", output.SerialNumber)
	assert.True(t, output.SmartStatus.Passed)
}

func TestSmartCtlOutput_WithNVMeFields(t *testing.T) {
	output := SmartCtlOutput{
		ModelName:              "Samsung SSD 980 PRO 1TB",
		SerialNumber:           "S5GXNF0N123456",
		FirmwareVersion:        "5B2QGXA7",
		NVMeControllerID:       0,
		NVMeIEEEOuiIdentifier:  0x002538,
		NVMeNumberOfNamespaces: 1,
		NVMeTotalCapacity:      1000204886016,
		NVMeUnallocatedCapacity: 0,
		NVMePCIVendor: &SmartCtlNVMePCIVendor{
			ID:          0x144D,
			SubsystemID: 0x144D,
		},
		NVMeSmartHealthInfoLog: &SmartCtlNVMeSmartHealthInfoLog{
			AvailableSpare: 100,
			PercentageUsed: 5,
			MediaErrors:    0,
		},
		NVMeNamespaces: []SmartCtlNVMENamespace{
			{
				ID: 1,
				Size: SmartCtlNVMeCapacity{
					Bytes: 1000204886016,
				},
			},
		},
	}

	assert.Equal(t, int64(1000204886016), output.NVMeTotalCapacity)
	assert.NotNil(t, output.NVMePCIVendor)
	assert.Equal(t, int64(0x144D), output.NVMePCIVendor.ID)
	assert.NotNil(t, output.NVMeSmartHealthInfoLog)
	assert.Equal(t, int64(5), output.NVMeSmartHealthInfoLog.PercentageUsed)
	assert.Len(t, output.NVMeNamespaces, 1)
}

func TestSmartCtlOutput_WithATAFields(t *testing.T) {
	output := SmartCtlOutput{
		ModelName:       "WDC WD10EZEX-00BN5A0",
		SerialNumber:    "WD-WMC3T0123456",
		FirmwareVersion: "01.01A01",
		ModelFamily:     "Western Digital Blue",
		RotationRate:    7200,
		ATAVersion: &SmartCtlATAVersion{
			MajorValue: 10,
			MinorValue: 0,
			String:     "ACS-3 T13/2161-D revision 5",
		},
		SATAVersion: &SmartCtlSATAVersion{
			String: "SATA 3.2",
			Value:  50,
		},
		ATASMARTAttributes: &SmartCtlATASMARTAttributes{
			Revision: 16,
			Table: []SmartCtlATASMARTEntry{
				{ID: 1, Name: "Raw_Read_Error_Rate", Value: 100},
				{ID: 5, Name: "Reallocated_Sector_Ct", Value: 100},
				{ID: 197, Name: "Current_Pending_Sector", Value: 100},
			},
		},
	}

	assert.Equal(t, "Western Digital Blue", output.ModelFamily)
	assert.Equal(t, int64(7200), output.RotationRate)
	assert.NotNil(t, output.ATAVersion)
	assert.NotNil(t, output.SATAVersion)
	assert.NotNil(t, output.ATASMARTAttributes)
	assert.Len(t, output.ATASMARTAttributes.Table, 3)
}

func TestSmartCtlOutput_WithSCSIFields(t *testing.T) {
	output := SmartCtlOutput{
		ModelName:       "SEAGATE ST4000NM0035",
		SerialNumber:    "ZC1234567890",
		FirmwareVersion: "TN03",
		SCSIModelName:   "ST4000NM0035",
		SCSIProduct:     "ST4000NM0035",
		SCSIVendor:      "SEAGATE",
		SCSIVersion:     "SPC-4",
		SCSIRevision:    "TN03",
		SCSIGrownDefectList: 5,
		SCSIProtectionType: 0,
		SCSIErrorCounterLog: &SmartCtlSCSIErrorCounterLog{
			Read: SmartCtlSCSIErrorDetails{
				TotalUncorrectedErrors: 0,
			},
		},
		SCSIStartStopCycleCounter: &SmartCtlSCSIStartStopCycle{
			AccumulatedStartStopCycles: 50,
		},
		SCSITransportProtocol: &SmartCtlSCSITransportProtocol{
			Name:  "SAS (SPL-4)",
			Value: 6,
		},
	}

	assert.Equal(t, "SEAGATE", output.SCSIVendor)
	assert.Equal(t, int64(5), output.SCSIGrownDefectList)
	assert.NotNil(t, output.SCSIErrorCounterLog)
	assert.NotNil(t, output.SCSIStartStopCycleCounter)
	assert.NotNil(t, output.SCSITransportProtocol)
}

func TestSmartCtlOutput_JSONDeserialization_NVMe(t *testing.T) {
	jsonData := `{
		"json_format_version": [1, 0],
		"smartctl": {
			"version": [7, 4],
			"exit_status": 0
		},
		"device": {
			"name": "/dev/nvme0n1",
			"info_name": "/dev/nvme0n1",
			"type": "nvme",
			"protocol": "NVMe"
		},
		"model_name": "Samsung SSD 980 PRO 1TB",
		"serial_number": "S5GXNF0N123456",
		"firmware_version": "5B2QGXA7",
		"nvme_pci_vendor": {
			"id": 5197,
			"subsystem_id": 5197
		},
		"nvme_total_capacity": 1000204886016,
		"nvme_smart_health_information_log": {
			"available_spare": 100,
			"available_spare_threshold": 10,
			"percentage_used": 5,
			"media_errors": 0,
			"num_err_log_entries": 0,
			"temperature": 35
		},
		"smart_status": {
			"passed": true
		},
		"smart_support": {
			"available": true,
			"enabled": true
		},
		"temperature": {
			"current": 35
		}
	}`

	var output SmartCtlOutput
	err := json.Unmarshal([]byte(jsonData), &output)
	assert.NoError(t, err)

	assert.Equal(t, "Samsung SSD 980 PRO 1TB", output.ModelName)
	assert.Equal(t, "/dev/nvme0n1", output.Device.Name)
	assert.Equal(t, "NVMe", output.Device.Protocol)
	assert.NotNil(t, output.NVMePCIVendor)
	assert.Equal(t, int64(5197), output.NVMePCIVendor.ID)
	assert.NotNil(t, output.NVMeSmartHealthInfoLog)
	assert.Equal(t, int64(5), output.NVMeSmartHealthInfoLog.PercentageUsed)
	assert.True(t, output.SmartStatus.Passed)
}

func TestSmartCtlOutput_DefaultValues(t *testing.T) {
	output := SmartCtlOutput{}

	assert.Empty(t, output.ModelName)
	assert.Empty(t, output.SerialNumber)
	assert.Empty(t, output.FirmwareVersion)
	assert.Nil(t, output.ATAVersion)
	assert.Nil(t, output.ATASMARTAttributes)
	assert.Nil(t, output.NVMePCIVendor)
	assert.Nil(t, output.NVMeSmartHealthInfoLog)
	assert.Nil(t, output.SCSIErrorCounterLog)
	assert.Equal(t, int64(0), output.RotationRate)
	assert.Equal(t, int64(0), output.PowerCycleCount)
}

func TestStringInfo_StructFields(t *testing.T) {
	info := StringInfo{
		String: "7.4-2024.01.01",
	}

	assert.Equal(t, "7.4-2024.01.01", info.String)
}

func TestSmartCtlNVMeCapacity_StructFields(t *testing.T) {
	capacity := SmartCtlNVMeCapacity{
		Blocks: 1953525168,
		Bytes:  1000204886016,
	}

	assert.Equal(t, int64(1953525168), capacity.Blocks)
	assert.Equal(t, int64(1000204886016), capacity.Bytes)
}

func TestSmartCtlNVMeEUI64_StructFields(t *testing.T) {
	eui := SmartCtlNVMeEUI64{
		ExtID: 0,
		OUI:   0x002538,
	}

	assert.Equal(t, int64(0), eui.ExtID)
	assert.Equal(t, int64(0x002538), eui.OUI)
}
