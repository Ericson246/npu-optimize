package hwinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVulkanGPU_AMD(t *testing.T) {
	output := `GPU id : 0 (AMD Radeon RX 7900 XTX)
Device Name   = AMD Radeon RX 7900 XTX`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "AMD Radeon RX 7900 XTX", name)
	assert.Equal(t, "amd", vendor)
}

func TestParseVulkanGPU_AMD_Lowercase(t *testing.T) {
	output := `GPU id : 0 (amd radeon rx 6800)
Device Name   = amd radeon rx 6800`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "amd radeon rx 6800", name)
	assert.Equal(t, "amd", vendor)
}

func TestParseVulkanGPU_Intel(t *testing.T) {
	output := `GPU id : 0 (Intel(R) Arc(TM) A770 Graphics)
Device Name   = Intel(R) Arc(TM) A770 Graphics`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "Intel(R) Arc(TM) A770 Graphics", name)
	assert.Equal(t, "intel", vendor)
}

func TestParseVulkanGPU_IntelIntegrated(t *testing.T) {
	output := `GPU id : 0 (Intel(R) UHD Graphics 770)
Device Name   = Intel(R) UHD Graphics 770`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "Intel(R) UHD Graphics 770", name)
	assert.Equal(t, "intel", vendor)
}

func TestParseVulkanGPU_NVIDIA(t *testing.T) {
	output := `GPU id : 0 (NVIDIA GeForce RTX 4090)
Device Name   = NVIDIA GeForce RTX 4090`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "NVIDIA GeForce RTX 4090", name)
	assert.Equal(t, "nvidia", vendor)
}

func TestParseVulkanGPU_Apple(t *testing.T) {
	output := `GPU id : 0 (Apple M3 Max)
Device Name   = Apple M3 Max`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "Apple M3 Max", name)
	assert.Equal(t, "apple", vendor)
}

func TestParseVulkanGPU_UnknownVendor(t *testing.T) {
	output := `GPU id : 0 (Custom GPU)
Device Name   = Custom GPU`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "Custom GPU", name)
	assert.Equal(t, "unknown", vendor)
}

func TestParseVulkanGPU_EmptyOutput(t *testing.T) {
	name, vendor := parseVulkanGPU("")
	assert.Equal(t, "", name)
	assert.Equal(t, "", vendor)
}

func TestParseVulkanGPU_NoGPUName(t *testing.T) {
	output := `Vulkan Instance Version: 1.3.275`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "", name)
	assert.Equal(t, "", vendor)
}

func TestParseVulkanGPU_ColonSeparator(t *testing.T) {
	output := `GPU id: 0 (AMD Radeon Graphics)
GPU name: AMD Radeon Graphics
Device Name: AMD Radeon Graphics`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "AMD Radeon Graphics", name)
	assert.Equal(t, "amd", vendor)
}

func TestParseVulkanGPU_ADeviceName(t *testing.T) {
	output := `GPU id : 0 (AMD Radeon Graphics)
Device Name   = AMD Radeon Graphics`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "AMD Radeon Graphics", name)
	assert.Equal(t, "amd", vendor)
}

func TestParseVulkanGPU_AdvancedMicroDevices(t *testing.T) {
	output := `Device Name   = Advanced Micro Devices, Inc. Radeon RX 7900 XT`
	name, vendor := parseVulkanGPU(output)
	assert.Equal(t, "Advanced Micro Devices, Inc. Radeon RX 7900 XT", name)
	assert.Equal(t, "amd", vendor)
}

func TestParseVulkanVRAM_DiscreteGPU(t *testing.T) {
	output := `Memory Heaps:
    memoryHeaps[0]:
      size = 8589934592
      flags: count = 1
        MEMORY_HEAP_DEVICE_LOCAL_BIT
    memoryHeaps[1]:
      size = 34359738368
      flags: count = 0`
	vram := parseVulkanVRAM(output)
	assert.Equal(t, int64(8192), vram)
}

func TestParseVulkanVRAM_OldFormat(t *testing.T) {
	output := `Memory Heaps:
  memoryHeaps[0]:
    size = 4294967296 (4 GiB)
    memoryHeapFlags: count = 1
        MEMORY_HEAP_DEVICE_LOCAL_BIT`
	vram := parseVulkanVRAM(output)
	assert.Equal(t, int64(4096), vram)
}

func TestParseVulkanVRAM_DiscreteWithSizeBeforeFlag(t *testing.T) {
	output := `Memory Heaps:
  memoryHeaps[0]:
    size = 21474836480
    memoryHeapFlags: count = 1
        MEMORY_HEAP_DEVICE_LOCAL_BIT`
	vram := parseVulkanVRAM(output)
	assert.Equal(t, int64(20480), vram)
}

func TestParseVulkanVRAM_NoDeviceLocal(t *testing.T) {
	output := `Memory Heaps:
    memoryHeaps[0]:
      size = 34359738368
      flags: count = 0`
	vram := parseVulkanVRAM(output)
	assert.Equal(t, int64(0), vram)
}

func TestParseVulkanVRAM_EmptyOutput(t *testing.T) {
	vram := parseVulkanVRAM("")
	assert.Equal(t, int64(0), vram)
}

func TestParseVulkanVRAM_NoMemorySection(t *testing.T) {
	output := `Vulkan Instance Version: 1.3.275
Device Properties:`
	vram := parseVulkanVRAM(output)
	assert.Equal(t, int64(0), vram)
}

func TestIsVulkanIntegrated_DiscreteGPU(t *testing.T) {
	output := `VkPhysicalDeviceProperties:
    deviceType = VK_PHYSICAL_DEVICE_TYPE_DISCRETE_GPU (2)
Memory Heaps:
    memoryHeaps[0]:
      size = 8589934592
      flags: count = 1
        MEMORY_HEAP_DEVICE_LOCAL_BIT`
	assert.False(t, isVulkanIntegrated(output))
}

func TestIsVulkanIntegrated_IntegratedGPU(t *testing.T) {
	output := `VkPhysicalDeviceProperties:
    deviceType = VK_PHYSICAL_DEVICE_TYPE_INTEGRATED_GPU (1)`
	assert.True(t, isVulkanIntegrated(output))
}

func TestIsVulkanIntegrated_NoDeviceLocalHeap(t *testing.T) {
	output := `Memory Heaps:
    memoryHeaps[0]:
      size = 34359738368
      flags: count = 0`
	assert.True(t, isVulkanIntegrated(output))
}

func TestIsVulkanIntegrated_DiscreteOldFormat(t *testing.T) {
	output := `Device Type: DISCRETE_GPU
Memory Heaps:
  memoryHeaps[0]:
    size = 8589934592
    memoryHeapFlags: count = 1
        MEMORY_HEAP_DEVICE_LOCAL_BIT`
	assert.False(t, isVulkanIntegrated(output))
}
