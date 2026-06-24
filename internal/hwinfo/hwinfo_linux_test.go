package hwinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCardDevice_Valid(t *testing.T) {
	assert.True(t, isCardDevice("card0"))
	assert.True(t, isCardDevice("card1"))
	assert.True(t, isCardDevice("card128"))
}

func TestIsCardDevice_Invalid(t *testing.T) {
	assert.False(t, isCardDevice("card0-HDMI-A-1"))
	assert.False(t, isCardDevice("card0-DVI-I-1"))
	assert.False(t, isCardDevice("renderD128"))
	assert.False(t, isCardDevice("card"))
	assert.False(t, isCardDevice("controlD64"))
}

func TestVendorMap_AMD(t *testing.T) {
	v, ok := vendorMap["0x1002"]
	assert.True(t, ok)
	assert.Equal(t, "amd", v)
}

func TestVendorMap_Intel(t *testing.T) {
	v, ok := vendorMap["0x8086"]
	assert.True(t, ok)
	assert.Equal(t, "intel", v)
}

func TestVendorMap_NVIDIA(t *testing.T) {
	v, ok := vendorMap["0x10de"]
	assert.True(t, ok)
	assert.Equal(t, "nvidia", v)
}

func TestVulkanDrivers_AMD(t *testing.T) {
	assert.True(t, vulkanDrivers["amdgpu"])
}

func TestVulkanDrivers_Intel(t *testing.T) {
	assert.True(t, vulkanDrivers["i915"])
	assert.True(t, vulkanDrivers["xe"])
}

func TestVulkanDrivers_NVIDIA(t *testing.T) {
	assert.True(t, vulkanDrivers["nvidia"])
	assert.True(t, vulkanDrivers["nouveau"])
}

func TestVulkanDrivers_Unsupported(t *testing.T) {
	assert.False(t, vulkanDrivers["efi-framebuffer"])
	assert.False(t, vulkanDrivers["simplefb"])
}

func TestPCIClassConstants(t *testing.T) {
	assert.Equal(t, "0x030000", pciClassVGA)
	assert.Equal(t, "0x038000", pciClassDisplay)
}
