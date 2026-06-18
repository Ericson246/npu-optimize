package hwinfo

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"

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
	_, err := os.Stat("/proc/driver/nvidia")
	return err == nil
}

func hasROCmRuntime() bool {
	_, err := os.Stat("/sys/class/kfd")
	if err != nil {
		return false
	}
	cmd := exec.Command("ldconfig", "-p")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "librocm")
}

func hasOpenVINORuntime() bool {
	dirs := []string{"/opt/intel/openvino", "/opt/intel/openvino_2026"}
	for _, dir := range dirs {
		_, err := os.Stat(dir)
		if err == nil {
			return true
		}
	}
	cmd := exec.Command("ldconfig", "-p")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "libopenvino.so")
}

func hasVulkanRuntime() bool {
	_, err := os.Stat("/usr/lib/x86_64-linux-gnu/libvulkan.so")
	if err == nil {
		return true
	}
	_, err = os.Stat("/usr/lib/aarch64-linux-gnu/libvulkan.so")
	return err == nil
}
