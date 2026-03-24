// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindVendor_DeviceModel(t *testing.T) {
	tests := []struct {
		name           string
		deviceModel    string
		modelFamily    string
		expectedVendor string
	}{
		// Seagate patterns
		{
			name:           "Seagate DL2400",
			deviceModel:    "DL2400MM0159",
			modelFamily:    "",
			expectedVendor: "Seagate",
		},
		{
			name:           "Seagate ST pattern",
			deviceModel:    "ST12000NM0008",
			modelFamily:    "",
			expectedVendor: "Seagate",
		},
		{
			name:           "Seagate ST1 pattern",
			deviceModel:    "ST1000NM0055",
			modelFamily:    "",
			expectedVendor: "Seagate",
		},
		{
			name:           "Seagate ST2 pattern",
			deviceModel:    "ST2000NM013A",
			modelFamily:    "",
			expectedVendor: "Seagate",
		},

		// Toshiba patterns
		{
			name:           "Toshiba explicit",
			deviceModel:    "TOSHIBA MG03ACA100",
			modelFamily:    "",
			expectedVendor: "Toshiba",
		},
		{
			name:           "Toshiba MG03 pattern",
			deviceModel:    "MG03ACA100",
			modelFamily:    "",
			expectedVendor: "Toshiba",
		},
		{
			name:           "Toshiba MG04 pattern",
			deviceModel:    "MG04ACA400N",
			modelFamily:    "",
			expectedVendor: "Toshiba",
		},
		{
			name:           "Toshiba MG06 pattern",
			deviceModel:    "MG06SCA800EY",
			modelFamily:    "",
			expectedVendor: "Toshiba",
		},
		{
			name:           "Toshiba MG08 pattern",
			deviceModel:    "MG08ADA400NY",
			modelFamily:    "",
			expectedVendor: "Toshiba",
		},

		// Intel patterns
		{
			name:           "Intel explicit",
			deviceModel:    "INTEL SSDSC2BB240G4",
			modelFamily:    "",
			expectedVendor: "Intel",
		},
		{
			name:           "Intel lowercase",
			deviceModel:    "intel ssdsc2kb480g8r",
			modelFamily:    "",
			expectedVendor: "Intel",
		},

		// Kioxia patterns
		{
			name:           "Kioxia explicit",
			deviceModel:    "KIOXIA KCD61LUL1T92",
			modelFamily:    "",
			expectedVendor: "Kioxia",
		},

		// Western Digital patterns
		{
			name:           "WDC explicit",
			deviceModel:    "WDC WD8004FRYZ-01VAEB0",
			modelFamily:    "",
			expectedVendor: "WesternDigital",
		},
		{
			name:           "Western Digital full name",
			deviceModel:    "Western Digital WD100EFAX",
			modelFamily:    "",
			expectedVendor: "WesternDigital",
		},
		{
			name:           "WD100 pattern",
			deviceModel:    "WD100EFAX-68LHPN0",
			modelFamily:    "",
			expectedVendor: "WesternDigital",
		},

		// HGST patterns
		{
			name:           "HGST explicit",
			deviceModel:    "HGST HUS726060ALE610",
			modelFamily:    "",
			expectedVendor: "HGST",
		},
		{
			name:           "HUS pattern",
			deviceModel:    "HUS722T1TALA600",
			modelFamily:    "",
			expectedVendor: "HGST",
		},
		{
			name:           "HUH pattern",
			deviceModel:    "HUH721010AL5200",
			modelFamily:    "",
			expectedVendor: "HGST",
		},

		// Micron patterns
		{
			name:           "Micron explicit",
			deviceModel:    "MICRON 5200_MTFDDAK960TDS",
			modelFamily:    "",
			expectedVendor: "Micron",
		},
		{
			name:           "MTFDD pattern",
			deviceModel:    "MTFDDAK960TDN",
			modelFamily:    "",
			expectedVendor: "Micron",
		},

		// SanDisk patterns
		{
			name:           "SanDisk explicit",
			deviceModel:    "SANDISK SDSSDA120G",
			modelFamily:    "",
			expectedVendor: "SanDisk",
		},

		// Samsung patterns
		{
			name:           "Samsung explicit",
			deviceModel:    "SAMSUNG MZ7LH480HBHQ",
			modelFamily:    "",
			expectedVendor: "Samsung",
		},
		{
			name:           "MZ7 pattern",
			deviceModel:    "MZ7LH480HBHQ0D3",
			modelFamily:    "",
			expectedVendor: "Samsung",
		},

		// Unknown vendor
		{
			name:           "Unknown vendor",
			deviceModel:    "UNKNOWN_RANDOM_MODEL",
			modelFamily:    "",
			expectedVendor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := FindVendor(tt.deviceModel, tt.modelFamily)
			assert.Equal(t, tt.expectedVendor, vendor)
		})
	}
}

func TestFindVendor_FromModelFamily(t *testing.T) {
	tests := []struct {
		name           string
		deviceModel    string
		modelFamily    string
		expectedVendor string
	}{
		{
			name:           "Intel from family",
			deviceModel:    "",
			modelFamily:    "Intel 520 Series SSDs",
			expectedVendor: "Intel",
		},
		{
			name:           "Seagate from family",
			deviceModel:    "",
			modelFamily:    "Seagate Exos X16",
			expectedVendor: "Seagate",
		},
		{
			name:           "Western Digital from family",
			deviceModel:    "",
			modelFamily:    "Western Digital Gold",
			expectedVendor: "WesternDigital",
		},
		{
			name:           "Toshiba from family",
			deviceModel:    "",
			modelFamily:    "Toshiba Enterprise Capacity MG07ACA Series",
			expectedVendor: "Toshiba",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := FindVendor(tt.deviceModel, tt.modelFamily)
			assert.Equal(t, tt.expectedVendor, vendor)
		})
	}
}

func TestFindVendor_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name           string
		deviceModel    string
		expectedVendor string
	}{
		{
			name:           "lowercase intel",
			deviceModel:    "intel ssdsc2bb240g4",
			expectedVendor: "Intel",
		},
		{
			name:           "mixed case seagate",
			deviceModel:    "SeAgAtE ST8000NM0055",
			expectedVendor: "Seagate",
		},
		{
			name:           "uppercase toshiba",
			deviceModel:    "TOSHIBA MG03ACA100",
			expectedVendor: "Toshiba",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor := FindVendor(tt.deviceModel, "")
			assert.Equal(t, tt.expectedVendor, vendor)
		})
	}
}

func TestFindVendor_DeviceModelTakesPrecedence(t *testing.T) {
	// When both device model and model family match different vendors,
	// the device model should be checked first
	vendor := FindVendor("INTEL SSDSC2BB240G4", "Seagate Enterprise")
	assert.Equal(t, "Intel", vendor)
}
