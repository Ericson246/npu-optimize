package backend

import "github.com/Ericson246/npu-optimize/internal/hwinfo"

type Type string

const (
	TypeLlamaCpp Type = "llama.cpp"
	TypeVLLM     Type = "vllm"
	TypeONNX     Type = "onnx"
)

type Interface interface {
	Type() Type
	Detect(hw *hwinfo.Info) bool
	Fit(modelPath string) (*FitResult, error)
	Benchmark(modelPath string, p Params) (*BenchResult, error)
	Sweep(modelPath string, baseline Params, mode string) ([]BenchResult, error)
}

type Params struct {
	NGPULayers int
	Threads    int
	NBatch     int
	NUBatch    int
	CtxSize    int
	FlashAttn  bool
	CacheTypeK string
	CacheTypeV string
	CPUMoE     bool
	NoMMAP     bool
	MLock      bool
	SpecType   *string
}

type FitResult struct {
	BuildCommit  string  `json:"build_commit"`
	ModelSize    int64   `json:"model_size"`
	NGPULayers   int     `json:"n_gpu_layers"`
	NBatch       int     `json:"n_batch"`
	NUBatch      int     `json:"n_ubatch"`
	NThreads     int     `json:"n_threads"`
	CtxSize      int     `json:"ctx_size"`
	FlashAttn    bool    `json:"flash_attn"`
	CacheTypeK   string  `json:"cache_type_k"`
	CacheTypeV   string  `json:"cache_type_v"`
	AvgTS        float64 `json:"avg_ts"`
	MaxTS        float64 `json:"max_ts"`
	BandwidthGBs float64 `json:"bandwidth_gbs"`
}

type BenchResult struct {
	Configuration Params
	AvgTS         float64
	MaxTS         float64
	BandwidthGBs  float64
	VRAMUsedMB    int64
}
