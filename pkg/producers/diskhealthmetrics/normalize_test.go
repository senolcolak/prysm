// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillDeviceInfoFromSmartData_ATA(t *testing.T) {
	smartData := &SmartCtlOutput{
		Device: SmartCtlDevice{
			Protocol: "ATA",
		},
		DeviceModel:     "WDC WD8004FRYZ-01VAEB0",
		SerialNumber:    "VDGE1234",
		FirmwareVersion: "01.01A01",
		ModelFamily:     "Western Digital Gold",
		RotationRate:    7200,
		FormFactor:      &SmartCtlFormFactor{Name: "3.5 inches"},
		SmartStatus: SmartCtlSmartStatus{
			Passed: true,
		},
	}

	deviceInfo := &DeviceInfo{}
	FillDeviceInfoFromSmartData(deviceInfo, smartData)

	assert.Equal(t, "WDC WD8004FRYZ-01VAEB0", deviceInfo.DeviceModel)
	assert.Equal(t, "VDGE1234", deviceInfo.SerialNumber)
	assert.Equal(t, "01.01A01", deviceInfo.FirmwareVersion)
	assert.Equal(t, "Western Digital Gold", deviceInfo.ModelFamily)
	assert.Equal(t, "ATA", deviceInfo.Vendor)
	assert.Equal(t, "WDC WD8004FRYZ-01VAEB0", deviceInfo.Product)
	assert.Equal(t, "hdd", deviceInfo.Media)
	assert.Equal(t, int64(7200), deviceInfo.RPM)
	assert.Equal(t, "3.5 inches", deviceInfo.FormFactor)
	assert.True(t, deviceInfo.HealthStatus)
}

func TestFillDeviceInfoFromSmartData_SCSI(t *testing.T) {
	userCapacity := &SmartCtlUserCapacity{Bytes: 8001563222016}
	smartData := &SmartCtlOutput{
		Device: SmartCtlDevice{
			Protocol: "SCSI",
			Type:     "scsi",
		},
		SCSIModelName: "SEAGATE ST8000NM0055",
		SCSIVendor:    "SEAGATE",
		SCSIProduct:   "ST8000NM0055",
		LogicalUnitID: "5000c500abcd1234",
		SerialNumber:  "ZA123456",
		UserCapacity:  userCapacity,
		RotationRate:  7200,
		FormFactor:    &SmartCtlFormFactor{Name: "3.5 inches"},
		SmartSupport: SmartCtlSmartSupport{
			Available: true,
		},
		SmartStatus: SmartCtlSmartStatus{
			Passed: true,
		},
	}

	deviceInfo := &DeviceInfo{}
	FillDeviceInfoFromSmartData(deviceInfo, smartData)

	assert.Equal(t, "SEAGATE ST8000NM0055", deviceInfo.DeviceModel)
	assert.Equal(t, "SEAGATE", deviceInfo.Vendor)
	assert.Equal(t, "ST8000NM0055", deviceInfo.Product)
	assert.Equal(t, "5000c500abcd1234", deviceInfo.LunID)
	assert.Equal(t, "hdd", deviceInfo.Media)
	assert.Equal(t, int64(7200), deviceInfo.RPM)
	assert.InDelta(t, 7452.0, deviceInfo.Capacity, 0.1) // ~8TB in GiB
	assert.True(t, deviceInfo.HealthStatus)
}

func TestFillDeviceInfoFromSmartData_NVMe(t *testing.T) {
	smartData := &SmartCtlOutput{
		Device: SmartCtlDevice{
			Protocol: "NVMe",
		},
		DeviceModel:       "Samsung SSD 970 EVO Plus 1TB",
		SerialNumber:      "S4EVNX0N123456",
		FirmwareVersion:   "2B2QEXM7",
		NVMeTotalCapacity: 1000204886016,
		NVMePCIVendor: &SmartCtlNVMePCIVendor{
			ID:          0x144D,
			SubsystemID: 0xA801,
		},
		NVMeSmartHealthInfoLog: &SmartCtlNVMeSmartHealthInfoLog{
			PercentageUsed: 5,
		},
		SmartStatus: SmartCtlSmartStatus{
			Passed: true,
		},
	}

	deviceInfo := &DeviceInfo{}
	FillDeviceInfoFromSmartData(deviceInfo, smartData)

	assert.Equal(t, "Samsung SSD 970 EVO Plus 1TB", deviceInfo.DeviceModel)
	assert.Equal(t, "S4EVNX0N123456", deviceInfo.SerialNumber)
	assert.Equal(t, "nvme", deviceInfo.Media)
	assert.Equal(t, int64(0), deviceInfo.RPM) // NVMe has no RPM
	assert.Contains(t, deviceInfo.Vendor, "0x144D")
	assert.Equal(t, "0x144D", deviceInfo.VendorID)
	assert.Equal(t, "0xA801", deviceInfo.SubsystemVendorID)
	assert.InDelta(t, 931.5, deviceInfo.Capacity, 0.1) // ~1TB in GiB
	assert.Equal(t, float64(5), deviceInfo.DWPD)
	assert.True(t, deviceInfo.HealthStatus)
}

func TestFillDeviceInfoFromSmartData_FailedHealth(t *testing.T) {
	smartData := &SmartCtlOutput{
		Device: SmartCtlDevice{
			Protocol: "ATA",
		},
		DeviceModel: "FAILING_DRIVE",
		SmartStatus: SmartCtlSmartStatus{
			Passed: false,
		},
	}

	deviceInfo := &DeviceInfo{}
	FillDeviceInfoFromSmartData(deviceInfo, smartData)

	assert.False(t, deviceInfo.HealthStatus)
}

func TestNormalizeDeviceInfo_IntelSSD(t *testing.T) {
	tests := []struct {
		name           string
		inputModel     string
		expectedVendor string
		expectedMedia  string
		expectedDWPD   float64
	}{
		{
			name:           "Intel S3610 200GB",
			inputModel:     "INTEL SSDSC2BX200G4R",
			expectedVendor: "Intel",
			expectedMedia:  "ssd",
			expectedDWPD:   3.0,
		},
		{
			name:           "Intel S4610 480GB",
			inputModel:     "SSDSC2KG240G8R",
			expectedVendor: "Intel",
			expectedMedia:  "ssd",
			expectedDWPD:   3.0,
		},
		{
			name:           "Intel S3500",
			inputModel:     "INTEL SSDSC2BB240G4",
			expectedVendor: "Intel",
			expectedMedia:  "ssd",
			expectedDWPD:   0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceInfo := &DeviceInfo{DeviceModel: tt.inputModel}
			NormalizeDeviceInfo(deviceInfo)

			assert.Equal(t, tt.expectedVendor, deviceInfo.Vendor)
			assert.Equal(t, tt.expectedMedia, deviceInfo.Media)
			assert.Equal(t, tt.expectedDWPD, deviceInfo.DWPD)
		})
	}
}

func TestNormalizeDeviceInfo_WesternDigital(t *testing.T) {
	tests := []struct {
		name           string
		inputModel     string
		expectedVendor string
		expectedMedia  string
		expectedRPM    int64
	}{
		{
			name:           "WD Gold 8TB",
			inputModel:     "WDC WD8004FRYZ-01VAEB0",
			expectedVendor: "WesternDigital",
			expectedMedia:  "hdd",
			expectedRPM:    7200,
		},
		{
			name:           "WD Red Plus",
			inputModel:     "WDC WD10JFCX-68N6GN0",
			expectedVendor: "WesternDigital",
			expectedMedia:  "hdd",
			expectedRPM:    5400,
		},
		{
			name:           "HGST Ultrastar",
			inputModel:     "HUS722T2TALA600",
			expectedVendor: "WesternDigital",
			expectedMedia:  "hdd",
			expectedRPM:    7200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceInfo := &DeviceInfo{DeviceModel: tt.inputModel}
			NormalizeDeviceInfo(deviceInfo)

			assert.Equal(t, tt.expectedVendor, deviceInfo.Vendor)
			assert.Equal(t, tt.expectedMedia, deviceInfo.Media)
			assert.Equal(t, tt.expectedRPM, deviceInfo.RPM)
		})
	}
}

func TestNormalizeDeviceInfo_Seagate(t *testing.T) {
	tests := []struct {
		name             string
		inputModel       string
		expectedVendor   string
		expectedProduct  string
		expectedCapacity float64
	}{
		{
			name:             "Seagate Exos 7E10 8TB",
			inputModel:       "ST8000NM014A",
			expectedVendor:   "Seagate",
			expectedProduct:  "Exos7E10",
			expectedCapacity: 8000,
		},
		{
			name:             "Seagate Exos 7E8 1TB",
			inputModel:       "ST1000NM0055-1V410C",
			expectedVendor:   "Seagate",
			expectedProduct:  "Exos7E8",
			expectedCapacity: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceInfo := &DeviceInfo{DeviceModel: tt.inputModel}
			NormalizeDeviceInfo(deviceInfo)

			assert.Equal(t, tt.expectedVendor, deviceInfo.Vendor)
			assert.Equal(t, tt.expectedProduct, deviceInfo.Product)
			assert.Equal(t, tt.expectedCapacity, deviceInfo.Capacity)
		})
	}
}

func TestNormalizeDeviceInfo_Toshiba(t *testing.T) {
	deviceInfo := &DeviceInfo{DeviceModel: "TOSHIBA MG03ACA100"}
	NormalizeDeviceInfo(deviceInfo)

	assert.Equal(t, "Toshiba", deviceInfo.Vendor)
	assert.Equal(t, "MG03", deviceInfo.Product)
	assert.Equal(t, float64(3000), deviceInfo.Capacity)
	assert.Equal(t, "hdd", deviceInfo.Media)
	assert.Equal(t, int64(7200), deviceInfo.RPM)
}

func TestNormalizeDeviceInfo_Samsung(t *testing.T) {
	deviceInfo := &DeviceInfo{DeviceModel: "MZ7LH480HBHQ0D3"}
	NormalizeDeviceInfo(deviceInfo)

	assert.Equal(t, "Samsung", deviceInfo.Vendor)
	assert.Equal(t, "PM883a", deviceInfo.Product)
	assert.Equal(t, float64(480), deviceInfo.Capacity)
	assert.Equal(t, "ssd", deviceInfo.Media)
	assert.Equal(t, 3.6, deviceInfo.DWPD)
}

func TestNormalizeDeviceInfo_Micron(t *testing.T) {
	deviceInfo := &DeviceInfo{DeviceModel: "MTFDDAK960TDN"}
	NormalizeDeviceInfo(deviceInfo)

	assert.Equal(t, "Micron", deviceInfo.Vendor)
	assert.Equal(t, "5200MAX", deviceInfo.Product)
	assert.Equal(t, float64(960), deviceInfo.Capacity)
	assert.Equal(t, "ssd", deviceInfo.Media)
	assert.Equal(t, 5.0, deviceInfo.DWPD)
}

func TestNormalizeDeviceInfo_UnknownModel(t *testing.T) {
	// Unknown models should not be modified
	deviceInfo := &DeviceInfo{
		DeviceModel: "UNKNOWN_MODEL_12345",
		Vendor:      "OriginalVendor",
	}
	NormalizeDeviceInfo(deviceInfo)

	assert.Equal(t, "OriginalVendor", deviceInfo.Vendor)
	assert.Equal(t, "UNKNOWN_MODEL_12345", deviceInfo.DeviceModel)
}

func TestNormalizeVendor_FromDeviceModel(t *testing.T) {
	tests := []struct {
		name           string
		deviceModel    string
		modelFamily    string
		expectedVendor string
	}{
		{
			name:           "Intel from model",
			deviceModel:    "INTEL SSDSC2BB240G4",
			modelFamily:    "",
			expectedVendor: "Intel",
		},
		{
			name:           "Toshiba from model",
			deviceModel:    "TOSHIBA MG03ACA100",
			modelFamily:    "",
			expectedVendor: "Toshiba",
		},
		{
			name:           "WDC prefix",
			deviceModel:    "WDC WD8004FRYZ",
			modelFamily:    "",
			expectedVendor: "WesternDigital",
		},
		{
			name:           "Seagate ST prefix",
			deviceModel:    "ST12000NM0008",
			modelFamily:    "",
			expectedVendor: "Seagate",
		},
		{
			name:           "HGST explicit",
			deviceModel:    "HGST HUS726060ALE610",
			modelFamily:    "",
			expectedVendor: "HGST",
		},
		{
			name:           "Samsung MZ7 prefix",
			deviceModel:    "MZ7LH480HBHQ0D3",
			modelFamily:    "",
			expectedVendor: "Samsung",
		},
		{
			name:           "Micron MTFD prefix",
			deviceModel:    "MTFDDAK960TDN",
			modelFamily:    "",
			expectedVendor: "Micron",
		},
		{
			name:           "Kioxia",
			deviceModel:    "KIOXIA KCD61LUL1T92",
			modelFamily:    "",
			expectedVendor: "Kioxia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceInfo := &DeviceInfo{
				DeviceModel: tt.deviceModel,
				ModelFamily: tt.modelFamily,
				Vendor:      "", // Empty to trigger normalization
			}
			NormalizeVendor(deviceInfo)

			assert.Equal(t, tt.expectedVendor, deviceInfo.Vendor)
		})
	}
}

func TestNormalizeVendor_FromModelFamily(t *testing.T) {
	deviceInfo := &DeviceInfo{
		DeviceModel: "UNKNOWN_MODEL",
		ModelFamily: "Western Digital Gold",
		Vendor:      "",
	}
	NormalizeVendor(deviceInfo)

	assert.Equal(t, "WesternDigital", deviceInfo.Vendor)
}

func TestNormalizeVendor_AlreadySet(t *testing.T) {
	deviceInfo := &DeviceInfo{
		DeviceModel: "INTEL SSDSC2BB240G4",
		Vendor:      "CustomVendor",
	}
	NormalizeVendor(deviceInfo)

	// Should not override existing vendor
	assert.Equal(t, "CustomVendor", deviceInfo.Vendor)
}

func TestNormalizeVendor_DL2400Seagate(t *testing.T) {
	deviceInfo := &DeviceInfo{
		DeviceModel: "DL2400MM0159",
		Vendor:      "",
	}
	NormalizeVendor(deviceInfo)

	assert.Equal(t, "Seagate", deviceInfo.Vendor)
}

func TestNormalizeVendor_MG0Toshiba(t *testing.T) {
	deviceInfo := &DeviceInfo{
		DeviceModel: "MG04ACA400N",
		Vendor:      "",
	}
	NormalizeVendor(deviceInfo)

	assert.Equal(t, "Toshiba", deviceInfo.Vendor)
}

func TestNormalizeDeviceInfo_DellNVMe(t *testing.T) {
	tests := []struct {
		name             string
		inputModel       string
		expectedProduct  string
		expectedCapacity float64
		expectedDWPD     float64
	}{
		{
			name:             "Dell P4610 1.6TB",
			inputModel:       "Dell Express Flash NVMe P4610 1.6TB SFF",
			expectedProduct:  "P4610-Dell",
			expectedCapacity: 1600,
			expectedDWPD:     3.0,
		},
		{
			name:             "Dell P4610 3.2TB",
			inputModel:       "Dell Express Flash NVMe P4610 3.2TB SFF",
			expectedProduct:  "P4610-Dell",
			expectedCapacity: 3200,
			expectedDWPD:     3.0,
		},
		{
			name:             "Dell P5600 3.2TB",
			inputModel:       "Dell Ent NVMe P5600 MU U.2 3.2TB",
			expectedProduct:  "P5600-Dell",
			expectedCapacity: 3200,
			expectedDWPD:     3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceInfo := &DeviceInfo{DeviceModel: tt.inputModel}
			NormalizeDeviceInfo(deviceInfo)

			assert.Equal(t, "Intel", deviceInfo.Vendor)
			assert.Equal(t, tt.expectedProduct, deviceInfo.Product)
			assert.Equal(t, tt.expectedCapacity, deviceInfo.Capacity)
			assert.Equal(t, tt.expectedDWPD, deviceInfo.DWPD)
			assert.Equal(t, "u2", deviceInfo.FormFactor)
		})
	}
}
