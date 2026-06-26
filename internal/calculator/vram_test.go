package calculator_test

import (
	"testing"

	calc "github.com/Ericson246/npu-optimize/internal/calculator"
	"github.com/stretchr/testify/assert"
)

type testHeader struct {
	nLayer     int
	nKVHeads   int
	nHeads     int
	hiddenSize int
	fileType   int
}

func (h *testHeader) NumLayer() int   { return h.nLayer }
func (h *testHeader) NumKVHeads() int { return h.nKVHeads }
func (h *testHeader) HeadDim() int {
	if h.nHeads == 0 || h.hiddenSize == 0 {
		return 128
	}
	return h.hiddenSize / h.nHeads
}
func (h *testHeader) QuantFileType() int { return h.fileType }

func TestCalculateVRAM_Fits(t *testing.T) {
	h := &testHeader{nLayer: 32, nKVHeads: 8, nHeads: 32, hiddenSize: 4096, fileType: 10}

	result := calc.CalculateVRAM(calc.Params{
		VRAMFreeMB: 8000,
		CtxSize:    4096,
		VRAMMargin: 400,
		FileSize:   4_000_000_000,
		Header:     h,
	})

	assert.True(t, result.FitsInVRAM)
	assert.Equal(t, int64(4_000_000_000), result.FileSizeBytes)
	assert.Greater(t, result.TotalBytes, int64(0))
	assert.Equal(t, -1, result.NGPULayers)
	assert.GreaterOrEqual(t, result.CtxMaxEstimate, 4096)
}

func TestCalculateVRAM_NoFit(t *testing.T) {
	h := &testHeader{nLayer: 32, nKVHeads: 8, nHeads: 32, hiddenSize: 4096, fileType: 10}

	result := calc.CalculateVRAM(calc.Params{
		VRAMFreeMB: 2000,
		CtxSize:    4096,
		VRAMMargin: 256,
		FileSize:   4_000_000_000,
		Header:     h,
	})

	assert.False(t, result.FitsInVRAM)
	assert.Equal(t, 0, result.NGPULayers)
}

func TestCalculateVRAM_NilHeader(t *testing.T) {
	result := calc.CalculateVRAM(calc.Params{
		VRAMFreeMB: 8000,
		CtxSize:    4096,
		VRAMMargin: 400,
		FileSize:   100_000_000,
		Header:     nil,
	})

	assert.True(t, result.FitsInVRAM)
	assert.Equal(t, int64(0), result.KVcacheBytes)
}

func TestCalculateVRAM_ZeroFileSize(t *testing.T) {
	h := &testHeader{nLayer: 32, nKVHeads: 8, nHeads: 32, hiddenSize: 4096, fileType: 10}

	result := calc.CalculateVRAM(calc.Params{
		VRAMFreeMB: 8000,
		CtxSize:    4096,
		VRAMMargin: 400,
		FileSize:   0,
		Header:     h,
	})

	assert.True(t, result.FitsInVRAM)
	assert.Equal(t, int64(0), result.FileSizeBytes)
}

func TestCalculateVRAM_CtxMaxLoop(t *testing.T) {
	h := &testHeader{nLayer: 1, nKVHeads: 1, nHeads: 1, hiddenSize: 64, fileType: 10}

	result := calc.CalculateVRAM(calc.Params{
		VRAMFreeMB: 50000,
		CtxSize:    4096,
		VRAMMargin: 1024,
		FileSize:   100_000_000,
		Header:     h,
	})

	assert.True(t, result.FitsInVRAM)
	assert.Greater(t, result.CtxMaxEstimate, 4096)
}

func TestCalculateVRAM_NoKVcache(t *testing.T) {
	h := &testHeader{nLayer: 0, nKVHeads: 0}

	result := calc.CalculateVRAM(calc.Params{
		VRAMFreeMB: 8000,
		CtxSize:    4096,
		VRAMMargin: 400,
		FileSize:   100_000_000,
		Header:     h,
	})

	assert.True(t, result.FitsInVRAM)
	assert.Equal(t, int64(0), result.KVcacheBytes)
}

func TestCalculateVRAM_TSestimate(t *testing.T) {
	h := &testHeader{nLayer: 1, nKVHeads: 1, nHeads: 1, hiddenSize: 64, fileType: 10}

	result := calc.CalculateVRAM(calc.Params{
		VRAMFreeMB:   8000,
		CtxSize:      4096,
		VRAMMargin:   1024,
		FileSize:     100_000_000,
		Header:       h,
		BandwidthGBs: 80.0,
	})

	assert.NotNil(t, result.TSEstimated)
	assert.Greater(t, *result.TSEstimated, float64(0))
}

func TestCalculateVRAM_NoTSestimateWithoutBandwidth(t *testing.T) {
	h := &testHeader{nLayer: 1, nKVHeads: 1, nHeads: 1, hiddenSize: 64, fileType: 10}

	result := calc.CalculateVRAM(calc.Params{
		VRAMFreeMB: 8000,
		CtxSize:    4096,
		VRAMMargin: 400,
		FileSize:   100_000_000,
		Header:     h,
	})

	assert.Nil(t, result.TSEstimated)
}
