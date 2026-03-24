// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and prysm contributors
//
// SPDX-License-Identifier: Apache-2.0

package diskhealthmetrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectOEMRelationship_Lenovo(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		model    string
		product  string
		expected string
	}{
		{
			name:     "Lenovo with Toshiba OEM",
			vendor:   "LENOVO",
			model:    "TOSHIBA MG04ACA200",
			product:  "",
			expected: "Lenovo (Toshiba OEM)",
		},
		{
			name:     "Lenovo with Seagate OEM via product",
			vendor:   "Lenovo",
			model:    "",
			product:  "Seagate ST8000NM0055",
			expected: "Lenovo (Seagate OEM)",
		},
		{
			name:     "Lenovo with HGST OEM",
			vendor:   "lenovo",
			model:    "HGST HUS726060ALE610",
			product:  "",
			expected: "Lenovo (HGST OEM)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectOEMRelationship(tt.vendor, tt.model, tt.product)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectOEMRelationship_Dell(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		model    string
		product  string
		expected string
	}{
		{
			name:     "Dell with Seagate OEM",
			vendor:   "Dell",
			model:    "Seagate ST12000NM0008",
			product:  "",
			expected: "Dell (Seagate OEM)",
		},
		{
			name:     "Dell with WD OEM via product",
			vendor:   "DELL",
			model:    "",
			product:  "Western Digital WD100EFAX",
			expected: "Dell (WD OEM)",
		},
		{
			name:     "Dell with WD OEM short name",
			vendor:   "Dell",
			model:    "",
			product:  "WD Red Plus",
			expected: "Dell (WD OEM)",
		},
		{
			name:     "Dell with Toshiba OEM",
			vendor:   "Dell",
			model:    "Toshiba MG07ACA14TE",
			product:  "",
			expected: "Dell (Toshiba OEM)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectOEMRelationship(tt.vendor, tt.model, tt.product)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectOEMRelationship_HP(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		model    string
		product  string
		expected string
	}{
		{
			name:     "HP with WD OEM",
			vendor:   "HP",
			model:    "Western Digital WD8004FRYZ",
			product:  "",
			expected: "HP (WD OEM)",
		},
		{
			name:     "HPE with Seagate OEM",
			vendor:   "HPE",
			model:    "",
			product:  "Seagate Enterprise",
			expected: "HP (Seagate OEM)",
		},
		{
			name:     "HP with Toshiba OEM",
			vendor:   "hp",
			model:    "Toshiba MG08ACA16TE",
			product:  "",
			expected: "HP (Toshiba OEM)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectOEMRelationship(tt.vendor, tt.model, tt.product)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectOEMRelationship_Supermicro(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		model    string
		product  string
		expected string
	}{
		{
			name:     "Supermicro with Intel OEM",
			vendor:   "Supermicro",
			model:    "Intel SSDSC2KB480G8",
			product:  "",
			expected: "Supermicro (Intel OEM)",
		},
		{
			name:     "Supermicro with Samsung OEM",
			vendor:   "SUPERMICRO",
			model:    "",
			product:  "Samsung PM883",
			expected: "Supermicro (Samsung OEM)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectOEMRelationship(tt.vendor, tt.model, tt.product)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectOEMRelationship_GenericOEM(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		model    string
		product  string
		expected string
	}{
		{
			name:     "Generic vendor with Seagate product",
			vendor:   "acme",
			model:    "",
			product:  "Seagate Exos X16",
			expected: "Acme (Seagate OEM)",
		},
		{
			name:     "Generic vendor with WD product",
			vendor:   "custom",
			model:    "",
			product:  "WD Gold",
			expected: "Custom (WD OEM)",
		},
		{
			name:     "Generic vendor with Toshiba product",
			vendor:   "server",
			model:    "",
			product:  "Toshiba Enterprise",
			expected: "Server (Toshiba OEM)",
		},
		{
			name:     "Generic vendor with HGST product",
			vendor:   "storage",
			model:    "",
			product:  "HGST Ultrastar",
			expected: "Storage (HGST OEM)",
		},
		{
			name:     "Generic vendor with Samsung product",
			vendor:   "nas",
			model:    "",
			product:  "Samsung PM883",
			expected: "Nas (Samsung OEM)",
		},
		{
			name:     "Generic vendor with Intel product",
			vendor:   "workstation",
			model:    "",
			product:  "Intel S4610",
			expected: "Workstation (Intel OEM)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectOEMRelationship(tt.vendor, tt.model, tt.product)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectOEMRelationship_NoOEM(t *testing.T) {
	tests := []struct {
		name    string
		vendor  string
		model   string
		product string
	}{
		{
			name:    "Direct Seagate",
			vendor:  "Seagate",
			model:   "ST8000NM0055",
			product: "",
		},
		{
			name:    "Direct Intel",
			vendor:  "Intel",
			model:   "SSDSC2KB480G8",
			product: "S4610",
		},
		{
			name:    "Direct Samsung",
			vendor:  "Samsung",
			model:   "MZ7LH480HAHQ",
			product: "PM883",
		},
		{
			name:    "Unknown vendor and product",
			vendor:  "Unknown",
			model:   "RANDOM_MODEL",
			product: "RANDOM_PRODUCT",
		},
		{
			name:    "Empty fields",
			vendor:  "",
			model:   "",
			product: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectOEMRelationship(tt.vendor, tt.model, tt.product)
			assert.Empty(t, result, "Should return empty for non-OEM or direct vendor")
		})
	}
}

func TestDetectOEMRelationship_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		vendor   string
		model    string
		product  string
		expected string
	}{
		{
			name:     "Lowercase vendor uppercase model",
			vendor:   "lenovo",
			model:    "TOSHIBA MG04ACA200",
			product:  "",
			expected: "Lenovo (Toshiba OEM)",
		},
		{
			name:     "Mixed case all fields",
			vendor:   "DeLL",
			model:    "SeAgAtE ST8000NM0055",
			product:  "",
			expected: "Dell (Seagate OEM)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectOEMRelationship(tt.vendor, tt.model, tt.product)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnhanceDeviceInfo_NilInput(t *testing.T) {
	// Should not panic with nil input
	assert.NotPanics(t, func() {
		enhanceDeviceInfo(nil)
	})
}

func TestEnhanceDeviceInfo_SetsModelFamily(t *testing.T) {
	deviceInfo := &DeviceInfo{
		Vendor:      "Lenovo",
		DeviceModel: "Toshiba MG04ACA200",
		Product:     "",
		ModelFamily: "", // Empty, should be set
	}

	enhanceDeviceInfo(deviceInfo)

	assert.Equal(t, "Lenovo (Toshiba OEM)", deviceInfo.ModelFamily)
}

func TestEnhanceDeviceInfo_PreservesExistingModelFamily(t *testing.T) {
	deviceInfo := &DeviceInfo{
		Vendor:      "Lenovo",
		DeviceModel: "Toshiba MG04ACA200",
		Product:     "",
		ModelFamily: "Existing Family", // Already set
	}

	enhanceDeviceInfo(deviceInfo)

	// Should not override existing ModelFamily
	assert.Equal(t, "Existing Family", deviceInfo.ModelFamily)
}

func TestEnhanceDeviceInfo_NoOEMRelationship(t *testing.T) {
	deviceInfo := &DeviceInfo{
		Vendor:      "Seagate",
		DeviceModel: "ST8000NM0055",
		Product:     "Exos 7E8",
		ModelFamily: "",
	}

	enhanceDeviceInfo(deviceInfo)

	// ModelFamily should remain empty when no OEM relationship detected
	assert.Empty(t, deviceInfo.ModelFamily)
}
