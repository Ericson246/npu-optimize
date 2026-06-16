package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppConstants(t *testing.T) {
	assert.Equal(t, "npu-optimize", AppName)
	assert.Equal(t, "0.1.1", Version)
	assert.Equal(t, "MIT", License)
	assert.Equal(t, "npu-optimize/0.1.1", UserAgent)
}

func TestDefaultValues(t *testing.T) {
	assert.Equal(t, 16384, DefaultCtxSize)
	assert.Equal(t, 1024, DefaultVRAMMargin)
	assert.Equal(t, 3.0, DefaultMinTS)
}

func TestHFConstants(t *testing.T) {
	assert.Equal(t, "https://huggingface.co", HFAPIBaseURL)
	assert.Equal(t, "huggingface.co", HFAPIHost)
}

func TestLlamaBench(t *testing.T) {
	assert.Equal(t, "b9180", LlamaBenchVersion)
	assert.Equal(t, "ggml-org/llama.cpp", LlamaBenchRepo)
}

func TestCacheDir(t *testing.T) {
	assert.Equal(t, ".npu-optimize", CacheDir)
	assert.Equal(t, "cache/hardware", CacheHardware)
}

func TestProxyModels(t *testing.T) {
	assert.Len(t, ProxyModels, 3)

	assert.Equal(t, "Qwen/Qwen2.5-0.5B-GGUF", ProxyModels[0].Repo)
	assert.Equal(t, "qwen2.5-0.5b-q4_k_m.gguf", ProxyModels[0].File)
	assert.Equal(t, int64(100_000_000), ProxyModels[0].Size)

	assert.Equal(t, "microsoft/Phi-3-mini-4k-instruct-gguf", ProxyModels[1].Repo)
	assert.Equal(t, "Phi-3-mini-4k-instruct-q4.gguf", ProxyModels[1].File)
	assert.Equal(t, int64(250_000_000), ProxyModels[1].Size)

	assert.Equal(t, "google/gemma-2-2b-it-GGUF", ProxyModels[2].Repo)
	assert.Equal(t, "gemma-2-2b-it-q4_k_m.gguf", ProxyModels[2].File)
	assert.Equal(t, int64(1_500_000_000), ProxyModels[2].Size)
}
