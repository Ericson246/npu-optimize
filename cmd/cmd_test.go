package cmd

import (
	"os"
	"testing"

	"github.com/Ericson246/npu-optimize/internal/hwinfo"
	"github.com/stretchr/testify/assert"
)

func nvidiaHW() *hwinfo.Info {
	return &hwinfo.Info{
		GPU: &hwinfo.GPUInfo{
			Vendor:      "nvidia",
			Name:        "RTX 4060",
			VRAMTotalMB: 8192,
			VRAMFreeMB:  6144,
			Integrated:  false,
		},
		CPU:        hwinfo.CPUInfo{Name: "CPU", Cores: 8, Threads: 16},
		RAMTotalMB: 32768,
		RAMFreeMB:  16384,
	}
}

func amdHW() *hwinfo.Info {
	return &hwinfo.Info{
		GPU: &hwinfo.GPUInfo{
			Vendor:      "amd",
			Name:        "AMD Radeon RX 7900 XTX",
			VRAMTotalMB: 24576,
			VRAMFreeMB:  22528,
			Integrated:  false,
		},
		CPU:        hwinfo.CPUInfo{Name: "CPU", Cores: 16, Threads: 32},
		RAMTotalMB: 65536,
		RAMFreeMB:  32768,
	}
}

func nvidiaLowVRAMHW() *hwinfo.Info {
	return &hwinfo.Info{
		GPU: &hwinfo.GPUInfo{
			Vendor:      "nvidia",
			Name:        "GTX 1650",
			VRAMTotalMB: 4096,
			VRAMFreeMB:  2048,
			Integrated:  false,
		},
		CPU:        hwinfo.CPUInfo{Name: "CPU", Cores: 4, Threads: 8},
		RAMTotalMB: 16384,
		RAMFreeMB:  8192,
	}
}

func integratedHW() *hwinfo.Info {
	return &hwinfo.Info{
		GPU: &hwinfo.GPUInfo{
			Vendor:      "intel",
			Name:        "UHD Graphics 770",
			VRAMTotalMB: 0,
			VRAMFreeMB:  0,
			Integrated:  true,
		},
		CPU:        hwinfo.CPUInfo{Name: "CPU", Cores: 8, Threads: 16},
		RAMTotalMB: 32768,
		RAMFreeMB:  16384,
	}
}

func cpuHW() *hwinfo.Info {
	return &hwinfo.Info{
		CPU:        hwinfo.CPUInfo{Name: "CPU Only", Cores: 8, Threads: 8},
		RAMTotalMB: 32768,
		RAMFreeMB:  16384,
	}
}

func lowRAMCPU() *hwinfo.Info {
	return &hwinfo.Info{
		CPU:        hwinfo.CPUInfo{Name: "Low RAM CPU", Cores: 2, Threads: 2},
		RAMTotalMB: 2048,
		RAMFreeMB:  512,
	}
}

func TestResolveDetectConfig_Auto_NVIDIA(t *testing.T) {
	cfg, err := resolveDetectConfig("auto", nvidiaHW())
	assert.NoError(t, err)
	assert.Equal(t, "gpu-only", cfg.modeUsed)
	assert.Equal(t, int64(6144), cfg.availableMemoryMB)
	assert.Equal(t, -1, cfg.nGPULayers)
	assert.Equal(t, 2048, cfg.nBatch)
	assert.True(t, cfg.flashAttn)
}

func TestResolveDetectConfig_Auto_Integrated(t *testing.T) {
	cfg, err := resolveDetectConfig("auto", integratedHW())
	assert.NoError(t, err)
	assert.Equal(t, "partial", cfg.modeUsed)
	assert.Equal(t, int64(0+16384*70/100), cfg.availableMemoryMB)
	assert.Equal(t, -1, cfg.nGPULayers)
	assert.Equal(t, 1024, cfg.nBatch)
	assert.False(t, cfg.flashAttn)
}

func TestResolveDetectConfig_Auto_CPU(t *testing.T) {
	cfg, err := resolveDetectConfig("auto", cpuHW())
	assert.NoError(t, err)
	assert.Equal(t, "cpu", cfg.modeUsed)
	assert.Equal(t, int64(16384*70/100), cfg.availableMemoryMB)
	assert.Equal(t, 0, cfg.nGPULayers)
	assert.Equal(t, 512, cfg.nBatch)
	assert.False(t, cfg.flashAttn)
}

func TestResolveDetectConfig_Auto_NoGPU_NoRAM(t *testing.T) {
	_, err := resolveDetectConfig("auto", lowRAMCPU())
	assert.Error(t, err)
	var hwErr *hwUnsupportedError
	assert.ErrorAs(t, err, &hwErr)
	assert.Contains(t, hwErr.Error(), "RAM")
}

func TestResolveDetectConfig_Auto_NVIDIA_LowVRAM(t *testing.T) {
	cfg, err := resolveDetectConfig("auto", nvidiaLowVRAMHW())
	assert.NoError(t, err)
	assert.Equal(t, "partial", cfg.modeUsed)
	assert.Equal(t, int64(2048+8192*70/100), cfg.availableMemoryMB)
	assert.Equal(t, -1, cfg.nGPULayers)
	assert.Equal(t, 1024, cfg.nBatch)
	assert.True(t, cfg.flashAttn)
}

func TestResolveDetectConfig_Auto_AMD(t *testing.T) {
	cfg, err := resolveDetectConfig("auto", amdHW())
	assert.NoError(t, err)
	assert.Equal(t, "gpu-only", cfg.modeUsed)
	assert.Equal(t, int64(22528), cfg.availableMemoryMB)
	assert.Equal(t, -1, cfg.nGPULayers)
	assert.True(t, cfg.flashAttn)
}

func TestResolveDetectConfig_GPUOnly_WithNVIDIA(t *testing.T) {
	cfg, err := resolveDetectConfig("gpu-only", nvidiaHW())
	assert.NoError(t, err)
	assert.Equal(t, "gpu-only", cfg.modeUsed)
	assert.Equal(t, int64(6144), cfg.availableMemoryMB)
	assert.True(t, cfg.flashAttn)
}

func TestResolveDetectConfig_GPUOnly_NoDiscreteGPU(t *testing.T) {
	_, err := resolveDetectConfig("gpu-only", cpuHW())
	assert.Error(t, err)
	var hwErr *hwUnsupportedError
	assert.ErrorAs(t, err, &hwErr)
	assert.Contains(t, hwErr.Error(), "discrete GPU")
}

func TestResolveDetectConfig_GPUOnly_LowVRAM(t *testing.T) {
	_, err := resolveDetectConfig("gpu-only", nvidiaLowVRAMHW())
	assert.Error(t, err)
	var hwErr *hwUnsupportedError
	assert.ErrorAs(t, err, &hwErr)
	assert.Contains(t, hwErr.Error(), "discrete GPU")
}

func TestResolveDetectConfig_GPUOnly_AMD(t *testing.T) {
	cfg, err := resolveDetectConfig("gpu-only", amdHW())
	assert.NoError(t, err)
	assert.Equal(t, "gpu-only", cfg.modeUsed)
	assert.Equal(t, int64(22528), cfg.availableMemoryMB)
	assert.Equal(t, -1, cfg.nGPULayers)
	assert.True(t, cfg.flashAttn)
}

func TestResolveDetectConfig_CPU_SufficientRAM(t *testing.T) {
	cfg, err := resolveDetectConfig("cpu", cpuHW())
	assert.NoError(t, err)
	assert.Equal(t, "cpu", cfg.modeUsed)
	assert.Equal(t, int64(16384*70/100), cfg.availableMemoryMB)
	assert.Equal(t, 0, cfg.nGPULayers)
	assert.Equal(t, 512, cfg.nBatch)
	assert.False(t, cfg.flashAttn)
}

func TestResolveDetectConfig_CPU_InsufficientRAM(t *testing.T) {
	_, err := resolveDetectConfig("cpu", lowRAMCPU())
	assert.Error(t, err)
	var hwErr *hwUnsupportedError
	assert.ErrorAs(t, err, &hwErr)
	assert.Contains(t, hwErr.Error(), "RAM")
}

func TestResolveDetectConfig_Partial(t *testing.T) {
	cfg, err := resolveDetectConfig("partial", nvidiaHW())
	assert.NoError(t, err)
	assert.Equal(t, "partial", cfg.modeUsed)
	assert.Equal(t, int64(6144+16384*70/100), cfg.availableMemoryMB)
	assert.Equal(t, -1, cfg.nGPULayers)
	assert.Equal(t, 1024, cfg.nBatch)
	assert.True(t, cfg.flashAttn)
}

func TestResolveDetectConfig_Partial_NoGPU_LowRAM(t *testing.T) {
	_, err := resolveDetectConfig("partial", lowRAMCPU())
	assert.Error(t, err)
}

func TestCalcVRAMMargin_High(t *testing.T) {
	assert.Equal(t, 1024, calcVRAMMargin(50000))
}

func TestCalcVRAMMargin_Normal(t *testing.T) {
	assert.Equal(t, 400, calcVRAMMargin(8000))
}

func TestCalcVRAMMargin_Low(t *testing.T) {
	assert.Equal(t, 256, calcVRAMMargin(1000))
}

func TestCalcVRAMMargin_Zero(t *testing.T) {
	assert.Equal(t, 256, calcVRAMMargin(0))
}

func TestGetToken_FromPackageVar(t *testing.T) {
	orig := token
	token = "test-token"
	defer func() { token = orig }()

	assert.Equal(t, "test-token", getToken())
}

func TestGetToken_FromEnvHF(t *testing.T) {
	orig := token
	token = ""
	defer func() { token = orig }()

	os.Setenv("HF_TOKEN", "hf-env-token")
	defer os.Unsetenv("HF_TOKEN")

	assert.Equal(t, "hf-env-token", getToken())
}

func TestGetToken_FromEnvNPU(t *testing.T) {
	orig := token
	token = ""
	defer func() { token = orig }()

	os.Setenv("NPU_OPTIMIZE_TOKEN", "npu-env-token")
	defer os.Unsetenv("NPU_OPTIMIZE_TOKEN")

	assert.Equal(t, "npu-env-token", getToken())
}

func TestGetToken_PackageVarOverridesEnv(t *testing.T) {
	orig := token
	token = "pkg-token"
	defer func() { token = orig }()

	os.Setenv("HF_TOKEN", "hf-env-token")
	defer os.Unsetenv("HF_TOKEN")

	assert.Equal(t, "pkg-token", getToken())
}

func TestGetToken_Empty(t *testing.T) {
	orig := token
	token = ""
	defer func() { token = orig }()

	os.Unsetenv("HF_TOKEN")
	os.Unsetenv("NPU_OPTIMIZE_TOKEN")

	assert.Empty(t, getToken())
}
