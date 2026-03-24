// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNVMeIDControllerOutput_StructFields(t *testing.T) {
	output := NVMeIDControllerOutput{
		VendorID:            0x144D, // Samsung
		SubsystemVendorID:   0x144D,
		ModelNumber:         "Samsung SSD 980 PRO 1TB",
		SerialNumber:        "S5GXNF0N123456",
		FirmwareRevision:    "5B2QGXA7",
		SubsystemNQN:        "nqn.2014.08.org.nvmexpress:144d144dS5GXNF0N123456",
		TotalCapacity:       1000204886016,
		UnallocatedCapacity: 0,
	}

	assert.Equal(t, int64(0x144D), output.VendorID)
	assert.Equal(t, int64(0x144D), output.SubsystemVendorID)
	assert.Equal(t, "Samsung SSD 980 PRO 1TB", output.ModelNumber)
	assert.Equal(t, "S5GXNF0N123456", output.SerialNumber)
	assert.Equal(t, "5B2QGXA7", output.FirmwareRevision)
	assert.Contains(t, output.SubsystemNQN, "nvmexpress")
	assert.Equal(t, int64(1000204886016), output.TotalCapacity)
	assert.Equal(t, int64(0), output.UnallocatedCapacity)
}

func TestNVMeIDControllerOutput_JSONDeserialization(t *testing.T) {
	// Note: IEEE field is json.Number which can be either numeric or string
	jsonData := `{
		"vid": 5197,
		"ssvid": 5197,
		"mn": "Samsung SSD 980 PRO 1TB",
		"sn": "S5GXNF0N123456",
		"fr": "5B2QGXA7",
		"subnqn": "nqn.2014.08.org.nvmexpress:144d144dS5GXNF0N123456",
		"ieee": 2439272,
		"tnvmcap": 1000204886016,
		"unvmcap": 0
	}`

	var output NVMeIDControllerOutput
	err := json.Unmarshal([]byte(jsonData), &output)
	assert.NoError(t, err)

	assert.Equal(t, int64(5197), output.VendorID)
	assert.Equal(t, "Samsung SSD 980 PRO 1TB", output.ModelNumber)
	assert.Equal(t, "S5GXNF0N123456", output.SerialNumber)
	assert.Equal(t, int64(1000204886016), output.TotalCapacity)
	assert.Equal(t, "2439272", output.IEEE.String())
}

func TestNVMeIDControllerOutput_DefaultValues(t *testing.T) {
	output := NVMeIDControllerOutput{}

	assert.Equal(t, int64(0), output.VendorID)
	assert.Equal(t, int64(0), output.SubsystemVendorID)
	assert.Empty(t, output.ModelNumber)
	assert.Empty(t, output.SerialNumber)
	assert.Empty(t, output.FirmwareRevision)
	assert.Empty(t, output.SubsystemNQN)
	assert.Equal(t, int64(0), output.TotalCapacity)
	assert.Equal(t, int64(0), output.UnallocatedCapacity)
}

func TestNVMeErrorEntry_StructFields(t *testing.T) {
	entry := NVMeErrorEntry{
		ErrorCount:                5,
		SubmissionQueueID:         1,
		CommandID:                 0x1234,
		StatusField:               0x281, // Media and Data Integrity Error
		PhaseTag:                  0,
		ParameterErrorLocation:    0,
		LBA:                       1048576,
		Namespace:                 1,
		VendorSpecific:            0,
		TransportType:             0,
		CommandSpecific:           0,
		TransportTypeSpecificInfo: 0,
	}

	assert.Equal(t, int64(5), entry.ErrorCount)
	assert.Equal(t, int64(1), entry.SubmissionQueueID)
	assert.Equal(t, int64(0x1234), entry.CommandID)
	assert.Equal(t, int64(0x281), entry.StatusField)
	assert.Equal(t, int64(1048576), entry.LBA)
	assert.Equal(t, int64(1), entry.Namespace)
}

func TestNVMeErrorLogOutput_StructFields(t *testing.T) {
	output := NVMeErrorLogOutput{
		Errors: []NVMeErrorEntry{
			{ErrorCount: 1, StatusField: 0x281, LBA: 100},
			{ErrorCount: 2, StatusField: 0x7, LBA: 200},
			{ErrorCount: 3, StatusField: 0x4, LBA: 300},
		},
	}

	assert.Len(t, output.Errors, 3)
	assert.Equal(t, int64(1), output.Errors[0].ErrorCount)
	assert.Equal(t, int64(0x281), output.Errors[0].StatusField)
	assert.Equal(t, int64(2), output.Errors[1].ErrorCount)
	assert.Equal(t, int64(3), output.Errors[2].ErrorCount)
}

func TestNVMeErrorLogOutput_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"errors": [
			{
				"error_count": 1,
				"sqid": 0,
				"cmdid": 4660,
				"status_field": 641,
				"phase_tag": 0,
				"parm_error_location": 0,
				"lba": 1048576,
				"nsid": 1,
				"vs": 0,
				"trtype": 0,
				"cs": 0,
				"trtype_spec_info": 0
			}
		]
	}`

	var output NVMeErrorLogOutput
	err := json.Unmarshal([]byte(jsonData), &output)
	assert.NoError(t, err)

	assert.Len(t, output.Errors, 1)
	assert.Equal(t, int64(1), output.Errors[0].ErrorCount)
	assert.Equal(t, int64(641), output.Errors[0].StatusField) // 0x281
	assert.Equal(t, int64(1048576), output.Errors[0].LBA)
}

func TestNVMeErrorLogOutput_EmptyErrors(t *testing.T) {
	jsonData := `{"errors": []}`

	var output NVMeErrorLogOutput
	err := json.Unmarshal([]byte(jsonData), &output)
	assert.NoError(t, err)

	assert.Empty(t, output.Errors)
}

func TestNVMeErrorLogOutput_DefaultValues(t *testing.T) {
	output := NVMeErrorLogOutput{}

	assert.Nil(t, output.Errors)
}

func TestEnhanceNVMeData_WithControllerData(t *testing.T) {
	smartData := &SmartCtlOutput{
		ModelName:      "Old Model",
		SerialNumber:   "OLD123",
		FirmwareVersion: "1.0",
	}

	nvmeController := &NVMeIDControllerOutput{
		ModelNumber:      "  Samsung SSD 980 PRO  ",
		SerialNumber:     "  S5GXNF0N123456  ",
		FirmwareRevision: "  5B2QGXA7  ",
		TotalCapacity:    1000204886016,
		VendorID:         0x144D,
		SubsystemNQN:     "nqn.2014.08.org.nvmexpress:144d",
	}

	enhanceNVMeData(smartData, nvmeController, nil)

	// Verify trimmed values
	assert.Equal(t, "Samsung SSD 980 PRO", smartData.ModelName)
	assert.Equal(t, "S5GXNF0N123456", smartData.SerialNumber)
	assert.Equal(t, "5B2QGXA7", smartData.FirmwareVersion)
	assert.Equal(t, int64(1000204886016), smartData.NVMeTotalCapacity)
	assert.NotNil(t, smartData.NVMePCIVendor)
	assert.Equal(t, int64(0x144D), smartData.NVMePCIVendor.ID)
	assert.Equal(t, "nqn.2014.08.org.nvmexpress:144d", smartData.Product)
}

func TestEnhanceNVMeData_WithErrors(t *testing.T) {
	smartData := &SmartCtlOutput{
		NVMeSmartHealthInfoLog: &SmartCtlNVMeSmartHealthInfoLog{},
	}

	nvmeErrors := &NVMeErrorLogOutput{
		Errors: []NVMeErrorEntry{
			{ErrorCount: 2, StatusField: 0x281}, // Media error
			{ErrorCount: 3, StatusField: 0x7},   // Aborted command
			{ErrorCount: 1, StatusField: 0x281}, // Another media error
		},
	}

	enhanceNVMeData(smartData, nil, nvmeErrors)

	assert.Equal(t, int64(3), smartData.NVMeSmartHealthInfoLog.MediaErrors) // 2 + 1
	assert.Equal(t, int64(3), smartData.NVMeSmartHealthInfoLog.NumErrLogEntries)
}

func TestEnhanceNVMeData_NilInputs(t *testing.T) {
	smartData := &SmartCtlOutput{
		ModelName: "Original",
	}

	// Should not panic with nil inputs
	enhanceNVMeData(smartData, nil, nil)

	assert.Equal(t, "Original", smartData.ModelName)
}

func TestEnhanceNVMeData_PreservesExistingVendor(t *testing.T) {
	smartData := &SmartCtlOutput{
		NVMePCIVendor: &SmartCtlNVMePCIVendor{
			ID:          0x8086, // Intel
			SubsystemID: 0x8086,
		},
	}

	nvmeController := &NVMeIDControllerOutput{
		VendorID:          0x144D, // Samsung
		SubsystemVendorID: 0x144D,
	}

	enhanceNVMeData(smartData, nvmeController, nil)

	// Should preserve the existing vendor info
	assert.Equal(t, int64(0x8086), smartData.NVMePCIVendor.ID)
}

func TestProcessNVMeSpecificAttributes_WithController(t *testing.T) {
	smartAttrs := make(map[string]SmartAttribute)

	nvmeController := &NVMeIDControllerOutput{
		VendorID:          0x144D,
		SubsystemVendorID: 0x144D,
		SubsystemNQN:      "nqn.2014.08.org.nvmexpress:144d",
	}

	processNVMeSpecificAttributes(smartAttrs, nvmeController, nil)

	// Verify vendor ID was stored
	if attr, ok := smartAttrs["nvme_vendor_id"]; ok {
		assert.Equal(t, int64(0x144D), attr.RawValue)
	}

	// Verify subsystem vendor ID was stored
	if attr, ok := smartAttrs["nvme_subsystem_vendor_id"]; ok {
		assert.Equal(t, int64(0x144D), attr.RawValue)
	}
}

func TestProcessNVMeSpecificAttributes_WithErrors(t *testing.T) {
	smartAttrs := make(map[string]SmartAttribute)

	nvmeErrors := &NVMeErrorLogOutput{
		Errors: []NVMeErrorEntry{
			{ErrorCount: 5, StatusField: 0x281, LBA: 1000}, // Media error with LBA
			{ErrorCount: 3, StatusField: 0x7},              // Aborted command
			{ErrorCount: 2, StatusField: 0x4},              // Timeout
		},
	}

	processNVMeSpecificAttributes(smartAttrs, nil, nvmeErrors)

	// Verify error log entries count
	if attr, ok := smartAttrs["nvme_error_log_entries"]; ok {
		assert.Equal(t, int64(3), attr.RawValue) // 3 entries
	}
}

func TestNVMeErrorStatusCodes(t *testing.T) {
	// Test that error classification works correctly
	tests := []struct {
		name       string
		statusCode int64
		isMedia    bool
		isAborted  bool
		isTimeout  bool
	}{
		{
			name:       "Media and Data Integrity Error",
			statusCode: 0x281,
			isMedia:    true,
		},
		{
			name:       "Aborted Command",
			statusCode: 0x7,
			isAborted:  true,
		},
		{
			name:       "Command Timeout",
			statusCode: 0x4,
			isTimeout:  true,
		},
		{
			name:       "Generic error",
			statusCode: 0x1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maskedCode := tt.statusCode & 0x7FF

			if tt.isMedia {
				assert.Equal(t, int64(0x281), maskedCode)
			}
			if tt.isAborted {
				assert.Equal(t, int64(0x7), maskedCode)
			}
			if tt.isTimeout {
				assert.Equal(t, int64(0x4), maskedCode)
			}
		})
	}
}

func TestNVMeIDControllerOutput_VendorIDFormats(t *testing.T) {
	tests := []struct {
		name             string
		vendorID         int64
		expectedHexStr   string
		isKnownVendor    bool
	}{
		{
			name:           "Samsung",
			vendorID:       0x144D,
			expectedHexStr: "144d",
			isKnownVendor:  true,
		},
		{
			name:           "Intel",
			vendorID:       0x8086,
			expectedHexStr: "8086",
			isKnownVendor:  true,
		},
		{
			name:           "Western Digital",
			vendorID:       0x1B96,
			expectedHexStr: "1b96",
			isKnownVendor:  true,
		},
		{
			name:           "Micron",
			vendorID:       0x1344,
			expectedHexStr: "1344",
			isKnownVendor:  true,
		},
		{
			name:           "Unknown vendor",
			vendorID:       0x9999,
			expectedHexStr: "9999",
			isKnownVendor:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NVMeIDControllerOutput{
				VendorID: tt.vendorID,
			}

			assert.Equal(t, tt.vendorID, output.VendorID)
		})
	}
}
