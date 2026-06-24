package hwinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseWMIJSON_SingleGPU(t *testing.T) {
	data := []byte(`{"Name":"NVIDIA GeForce RTX 4090","AdapterRAM":25769803776}`)
	gpu, ok := parseWMIJSON(data)
	assert.True(t, ok)
	assert.Equal(t, "nvidia", gpu.Vendor)
	assert.Equal(t, "NVIDIA GeForce RTX 4090", gpu.Name)
	assert.Equal(t, int64(24576), gpu.VRAMTotalMB)
	assert.False(t, gpu.Integrated)
}

func TestParseWMIJSON_MultipleGPUs(t *testing.T) {
	data := []byte(`[
		{"Name":"AMD Radeon RX 7900 XTX","AdapterRAM":25769803776},
		{"Name":"Intel(R) UHD Graphics 770","AdapterRAM":1073741824}
	]`)
	gpu, ok := parseWMIJSON(data)
	assert.True(t, ok)
	assert.Equal(t, "amd", gpu.Vendor)
	assert.Equal(t, "AMD Radeon RX 7900 XTX", gpu.Name)
	assert.Equal(t, int64(24576), gpu.VRAMTotalMB)
	assert.False(t, gpu.Integrated)
}

func TestParseWMIJSON_FiltersBasicDisplay(t *testing.T) {
	data := []byte(`[
		{"Name":"Microsoft Basic Display Adapter","AdapterRAM":0},
		{"Name":"AMD Radeon RX 6600","AdapterRAM":8589934592}
	]`)
	gpu, ok := parseWMIJSON(data)
	assert.True(t, ok)
	assert.Equal(t, "amd", gpu.Vendor)
	assert.Equal(t, "AMD Radeon RX 6600", gpu.Name)
}

func TestParseWMIJSON_Empty(t *testing.T) {
	_, ok := parseWMIJSON([]byte(""))
	assert.False(t, ok)
}

func TestParseWMIJSON_Null(t *testing.T) {
	_, ok := parseWMIJSON([]byte("null"))
	assert.False(t, ok)
}

func TestParseWMIJSON_IntelIntegrated(t *testing.T) {
	data := []byte(`{"Name":"Intel(R) UHD Graphics 770","AdapterRAM":1073741824}`)
	gpu, ok := parseWMIJSON(data)
	assert.True(t, ok)
	assert.Equal(t, "intel", gpu.Vendor)
	assert.True(t, gpu.Integrated)
	assert.Equal(t, int64(1024), gpu.VRAMTotalMB)
}

func TestBuildFromWMI_AMD(t *testing.T) {
	gpu := buildFromWMI(wmiGPU{Name: "AMD Radeon RX 6800", AdapterRAM: 17179869184})
	assert.Equal(t, "amd", gpu.Vendor)
	assert.Equal(t, int64(16384), gpu.VRAMTotalMB)
	assert.False(t, gpu.Integrated)
}

func TestBuildFromWMI_NVIDIA(t *testing.T) {
	gpu := buildFromWMI(wmiGPU{Name: "NVIDIA GeForce RTX 4060", AdapterRAM: 8589934592})
	assert.Equal(t, "nvidia", gpu.Vendor)
	assert.False(t, gpu.Integrated)
}

func TestBuildFromWMI_Intel(t *testing.T) {
	gpu := buildFromWMI(wmiGPU{Name: "Intel(R) Arc(TM) A770", AdapterRAM: 17179869184})
	assert.Equal(t, "intel", gpu.Vendor)
	assert.True(t, gpu.Integrated)
}

func TestBuildFromWMI_Unknown(t *testing.T) {
	gpu := buildFromWMI(wmiGPU{Name: "Custom GPU", AdapterRAM: 4096})
	assert.Equal(t, "unknown", gpu.Vendor)
	assert.False(t, gpu.Integrated)
}

func TestBuildFromWMI_Apple(t *testing.T) {
	gpu := buildFromWMI(wmiGPU{Name: "Apple M3 Max", AdapterRAM: 0})
	assert.Equal(t, "apple", gpu.Vendor)
	assert.True(t, gpu.Integrated)
}

func TestCUDALibs_Data(t *testing.T) {
	assert.Greater(t, len(cudaLibs), 0)
	for _, l := range cudaLibs {
		assert.Contains(t, l.name, ".dll")
		assert.NotEmpty(t, l.version)
	}
}

func TestROCmLibs_Data(t *testing.T) {
	assert.Greater(t, len(rocmLibs), 0)
	for _, l := range rocmLibs {
		assert.Contains(t, l.name, ".dll")
		assert.NotEmpty(t, l.version)
	}
}
