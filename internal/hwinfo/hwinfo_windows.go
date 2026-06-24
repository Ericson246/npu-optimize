package hwinfo

import (
	"encoding/json"
	"log/slog"
	"os/exec"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func detect() (*Info, error) {
	info := &Info{}
	detectCPU(info)
	detectRAM(info)
	detectGPU(info)
	detectBackends(info)
	return info, nil
}

func detectCPU(info *Info) {
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CPU.Name = cpuInfo[0].ModelName
	}
	info.CPU.Cores, _ = cpu.Counts(false)
	info.CPU.Threads, _ = cpu.Counts(true)
}

func detectRAM(info *Info) {
	vmem, err := mem.VirtualMemory()
	if err == nil {
		info.RAMTotalMB = int64(vmem.Total / 1024 / 1024)
		info.RAMFreeMB = int64(vmem.Available / 1024 / 1024)
	}
}

func detectBackends(info *Info) {
	backends := []string{}

	if hasCUDARuntime() {
		backends = append(backends, "cuda")
	}
	if hasROCmRuntime() && info.GPU != nil && info.GPU.Vendor == "amd" {
		backends = append(backends, "rocm")
	}
	if hasOpenVINORuntime() {
		backends = append(backends, "openvino")
	}
	if hasVulkanRuntime() {
		backends = append(backends, "vulkan")
	}

	if len(backends) == 0 {
		backends = append(backends, "cpu")
	}

	if info.GPU != nil {
		info.GPU.Backends = backends
	}

	slog.Debug("detected backends", "backends", backends)
}

func hasCUDARuntime() bool {
	for _, name := range []string{"cudart64_12.dll", "cudart64_13.dll", "cudart64_11.dll"} {
		lib, err := syscall.LoadLibrary(name)
		if err == nil {
			_ = syscall.FreeLibrary(lib)
			return true
		}
	}
	return false
}

func hasROCmRuntime() bool {
	for _, name := range []string{"amdhip64_7.dll", "amdhip64_6.dll"} {
		lib, err := syscall.LoadLibrary(name)
		if err == nil {
			_ = syscall.FreeLibrary(lib)
			return true
		}
	}
	return false
}

func hasOpenVINORuntime() bool {
	lib, err := syscall.LoadLibrary("openvino.dll")
	if err != nil {
		return false
	}
	_ = syscall.FreeLibrary(lib)
	return true
}

func hasVulkanRuntime() bool {
	lib, err := syscall.LoadLibrary("vulkan-1.dll")
	if err != nil {
		return false
	}
	_ = syscall.FreeLibrary(lib)
	return true
}

func detectVulkanGPUFallback(info *Info) bool {
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-CimInstance Win32_VideoController | Select-Object Name, AdapterRAM | ConvertTo-Json -Compress")
	out, err := cmd.Output()
	if err != nil {
		slog.Debug("vulkan fallback WMI command failed", "err", err)
		return false
	}

	gpu, ok := parseWMIJSON(out)
	if !ok {
		return false
	}

	gpu.VRAMFreeMB = gpu.VRAMTotalMB
	if gpu.Integrated && info.RAMFreeMB < gpu.VRAMTotalMB {
		gpu.VRAMFreeMB = info.RAMFreeMB
	}

	if gpu.VRAMTotalMB <= 0 {
		gpu.VRAMTotalMB = info.RAMTotalMB / 2
		gpu.Integrated = true
		gpu.VRAMFreeMB = info.RAMFreeMB
	}

	info.GPU = gpu
	slog.Warn("vulkan GPU detected via WMI fallback (vulkaninfo not found)",
		"vendor", gpu.Vendor, "gpu", gpu.Name,
		"vram_mb", gpu.VRAMTotalMB,
	)
	return true
}

type wmiGPU struct {
	Name       string `json:"Name"`
	AdapterRAM int64  `json:"AdapterRAM"`
}

func parseWMIJSON(data []byte) (*GPUInfo, bool) {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" {
		return nil, false
	}

	var single wmiGPU
	if err := json.Unmarshal(data, &single); err == nil && single.Name != "" {
		return buildFromWMI(single), true
	}

	var multiple []wmiGPU
	if err := json.Unmarshal(data, &multiple); err != nil {
		return nil, false
	}

	for _, g := range multiple {
		if g.Name == "" {
			continue
		}
		if !strings.Contains(strings.ToLower(g.Name), "microsoft basic display") {
			return buildFromWMI(g), true
		}
	}

	for _, g := range multiple {
		if g.Name != "" {
			return buildFromWMI(g), true
		}
	}

	return nil, false
}

func buildFromWMI(g wmiGPU) *GPUInfo {
	lower := strings.ToLower(g.Name)
	vendor := "unknown"
	switch {
	case strings.Contains(lower, "nvidia"):
		vendor = "nvidia"
	case strings.Contains(lower, "advanced micro devices"), strings.Contains(lower, "amd"), strings.Contains(lower, "radeon"):
		vendor = "amd"
	case strings.Contains(lower, "intel"):
		vendor = "intel"
	case strings.Contains(lower, "apple"):
		vendor = "apple"
	}

	integrated := vendor == "intel" || vendor == "apple"

	vramMB := g.AdapterRAM / 1024 / 1024

	return &GPUInfo{
		Vendor:      vendor,
		Name:        g.Name,
		VRAMTotalMB: vramMB,
		Integrated:  integrated,
	}
}
