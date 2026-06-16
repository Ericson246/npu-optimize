package calculator

type Header interface {
	NumLayer() int
	NumKVHeads() int
	HeadDim() int
	QuantFileType() int
}

type VRAMResult struct {
	FileSizeBytes  int64    `json:"file_size_bytes"`
	KVcacheBytes   int64    `json:"kv_cache_bytes"`
	OverheadBytes  int64    `json:"overhead_bytes"`
	TotalBytes     int64    `json:"total_bytes"`
	FitsInVRAM     bool     `json:"fits_in_vram"`
	NGPULayers     int      `json:"n_gpu_layers"`
	CtxMaxEstimate int      `json:"ctx_max_estimate"`
	TSEstimated    *float64 `json:"ts_estimated,omitempty"`
}

type Params struct {
	VRAMFreeMB   int64
	CtxSize      int
	VRAMMargin   int
	FileSize     int64
	Header       Header
	BandwidthGBs float64
}

func CalculateVRAM(p Params) *VRAMResult {
	vramFreeBytes := p.VRAMFreeMB * 1024 * 1024
	marginBytes := int64(p.VRAMMargin) * 1024 * 1024
	fileSize := p.FileSize

	h := p.Header

	var kvCacheBytes int64
	if h != nil && h.NumLayer() > 0 && h.NumKVHeads() > 0 {
		kvCacheBytes = int64(h.NumLayer()) * int64(h.NumKVHeads()) * int64(h.HeadDim()) * 2 * int64(p.CtxSize)
	}

	overheadBytes := marginBytes
	if h != nil {
		overheadBytes += int64(h.NumLayer()) * 10 * 1024 * 1024
	}
	totalBytes := fileSize + kvCacheBytes + overheadBytes

	fits := totalBytes <= vramFreeBytes

	nGPULayers := -1
	if !fits {
		nGPULayers = 0
	}

	ctxMax := p.CtxSize
	if fits && h != nil {
		for ctxMax < 131072 {
			newKVCache := int64(h.NumLayer()) * int64(h.NumKVHeads()) * int64(h.HeadDim()) * 2 * int64(ctxMax+4096)
			newTotal := fileSize + newKVCache + overheadBytes
			if newTotal <= vramFreeBytes {
				ctxMax += 4096
			} else {
				break
			}
		}
	}

	var ts *float64
	if p.BandwidthGBs > 0 && fileSize > 0 {
		bytesPerToken := float64(fileSize+int64(p.CtxSize)*100) / float64(p.CtxSize)
		estimated := p.BandwidthGBs / (bytesPerToken / 1e9)
		ts = &estimated
	}

	return &VRAMResult{
		FileSizeBytes:  fileSize,
		KVcacheBytes:   kvCacheBytes,
		OverheadBytes:  overheadBytes,
		TotalBytes:     totalBytes,
		FitsInVRAM:     fits,
		NGPULayers:     nGPULayers,
		CtxMaxEstimate: ctxMax,
		TSEstimated:    ts,
	}
}
