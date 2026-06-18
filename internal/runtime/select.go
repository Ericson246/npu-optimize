package runtime

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/Ericson246/npu-optimize/internal/hwinfo"
)

type Backend int

const (
	BackendUnknown Backend = iota
	BackendCUDA
	BackendROCm
	BackendOpenVINO
	BackendVulkanDiscrete
	BackendVulkanIntegrated
	BackendOpenVINONPU
	BackendCPU
)

func backendFromString(s string) Backend {
	switch strings.ToLower(s) {
	case "cuda":
		return BackendCUDA
	case "rocm":
		return BackendROCm
	case "openvino":
		return BackendOpenVINO
	case "vulkan":
		return BackendVulkanDiscrete
	case "vulkan-integrated":
		return BackendVulkanIntegrated
	case "vulkan-npu":
		return BackendOpenVINONPU
	case "cpu":
		return BackendCPU
	default:
		return BackendUnknown
	}
}

func Select(hw *hwinfo.Info, prefer string, catalog *Catalog) (*RuntimeEntry, error) {
	platform := runtime.GOOS
	arch := runtime.GOARCH

	archStr := "x64"
	if strings.Contains(arch, "arm64") || strings.Contains(arch, "aarch64") {
		archStr = "arm64"
	}

	backends := priorityBackends(hw, prefer)

	for _, b := range backends {
		entry := findRuntime(catalog, platform, archStr, b)
		if entry != nil {
			return entry, nil
		}
	}

	for _, src := range catalog.Sources {
		for id, entry := range src.Runtimes {
			if entry.Platform == platform && entry.Arch == archStr {
				entry.ID = id
				return &entry, nil
			}
		}
	}

	return nil, fmt.Errorf("no compatible runtime found for %s/%s", platform, arch)
}

func priorityBackends(hw *hwinfo.Info, prefer string) []Backend {
	if prefer != "" {
		b := backendFromString(prefer)
		if b != BackendUnknown {
			return []Backend{b}
		}
	}

	available := map[Backend]bool{}
	if hw.GPU != nil {
		for _, name := range hw.GPU.Backends {
			switch name {
			case "cuda":
				available[BackendCUDA] = true
			case "rocm":
				available[BackendROCm] = true
			case "openvino":
				available[BackendOpenVINO] = true
				available[BackendOpenVINONPU] = true
			case "vulkan":
				if hw.GPU.Integrated {
					available[BackendVulkanIntegrated] = true
				} else {
					available[BackendVulkanDiscrete] = true
				}
			}
		}
	}
	available[BackendCPU] = true

	priority := []Backend{
		BackendCUDA,
		BackendROCm,
		BackendOpenVINO,
		BackendVulkanDiscrete,
		BackendVulkanIntegrated,
		BackendOpenVINONPU,
		BackendCPU,
	}

	var result []Backend
	for _, b := range priority {
		if available[b] {
			result = append(result, b)
		}
	}

	if len(result) == 0 {
		result = append(result, BackendCPU)
	}

	return result
}

func findRuntime(catalog *Catalog, platform, arch string, backend Backend) *RuntimeEntry {
	backendStr := backendString(backend)
	if backendStr == "" {
		return nil
	}

	for _, src := range catalog.Sources {
		for id, entry := range src.Runtimes {
			if entry.Platform != platform || entry.Arch != arch {
				continue
			}
			if !strings.HasPrefix(entry.Backend, backendStr) && !strings.Contains(entry.ID, backendStr) {
				continue
			}
			entry.ID = id
			return &entry
		}
	}

	return nil
}

func backendString(b Backend) string {
	switch b {
	case BackendCUDA:
		return "cuda"
	case BackendROCm:
		return "rocm"
	case BackendOpenVINO:
		return "openvino"
	case BackendVulkanDiscrete:
		return "vulkan"
	case BackendVulkanIntegrated:
		return "vulkan"
	case BackendOpenVINONPU:
		return "openvino"
	case BackendCPU:
		return "cpu"
	default:
		return ""
	}
}
