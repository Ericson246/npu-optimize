package hwinfo

type GPUInfo struct {
	Vendor      string   `json:"vendor"`
	Name        string   `json:"name"`
	VRAMTotalMB int64    `json:"vram_total_mb"`
	VRAMFreeMB  int64    `json:"vram_free_mb"`
	Integrated  bool     `json:"integrated"`
	Backends    []string `json:"backends,omitempty"`
}

type CPUInfo struct {
	Name    string   `json:"name"`
	Cores   int      `json:"cores"`
	Threads int      `json:"threads"`
	ISA     []string `json:"isa,omitempty"`
}

type Info struct {
	GPU        *GPUInfo `json:"gpu,omitempty"`
	CPU        CPUInfo  `json:"cpu"`
	RAMTotalMB int64    `json:"ram_total_mb"`
	RAMFreeMB  int64    `json:"ram_free_mb"`
}

type Detector interface {
	Detect() (*Info, error)
}

type DetectorFunc func() (*Info, error)

func (f DetectorFunc) Detect() (*Info, error) {
	return f()
}

func Detect() (*Info, error) {
	return detect()
}

var DefaultDetector Detector = DetectorFunc(detect)
