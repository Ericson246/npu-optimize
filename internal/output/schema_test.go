package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	o := New(1)
	assert.Equal(t, 1, o.Version)
	assert.Equal(t, "llama.cpp", o.Backend)
	assert.Contains(t, o.Schema, "v1.json")
	assert.NotEmpty(t, o.GeneratedAt)
	assert.NotEmpty(t, o.ToolVersion)
	assert.False(t, o.Viable)
}

func TestNew_Version2(t *testing.T) {
	o := New(2)
	assert.Equal(t, 2, o.Version)
	assert.Contains(t, o.Schema, "v2.json")
}

func TestEncode_ValidJSON(t *testing.T) {
	o := New(1)
	o.ModeUsed = "gpu-only"
	o.Viable = true
	o.Hardware = &HardwareInfo{
		CPU:        CPUInfo{Name: "Test CPU", Cores: 8, Threads: 16},
		RAMTotalMB: 32768,
		RAMFreeMB:  16384,
	}

	var buf bytes.Buffer
	err := Encode(&buf, o)
	require.NoError(t, err)

	var decoded Output
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, 1, decoded.Version)
	assert.Equal(t, "gpu-only", decoded.ModeUsed)
	assert.True(t, decoded.Viable)
	assert.Equal(t, "llama.cpp", decoded.Backend)
}

func TestEncode_WithRecommended(t *testing.T) {
	o := New(1)
	o.Viable = true
	o.Recommended = &Recommended{
		Repo:       "test/repo",
		File:       "model.gguf",
		SizeBytes:  4_000_000_000,
		NLayers:    32,
		NKVHeads:   8,
		HeadDim:    128,
		FitsInVRAM: true,
	}

	var buf bytes.Buffer
	err := Encode(&buf, o)
	require.NoError(t, err)

	var decoded Output
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	require.NotNil(t, decoded.Recommended)
	assert.Equal(t, "test/repo", decoded.Recommended.Repo)
}

func TestEncodeError(t *testing.T) {
	var buf bytes.Buffer
	err := EncodeError(&buf, 4, "auth_required", "token needed", map[string]any{"endpoint": "/api/models"})
	require.NoError(t, err)

	var decoded ErrorOutput
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, 1, decoded.Version)
	assert.True(t, decoded.Error)
	assert.Equal(t, 4, decoded.ErrorCode)
	assert.Equal(t, "auth_required", decoded.ErrorType)
	assert.Equal(t, "token needed", decoded.Message)
	require.NotNil(t, decoded.Details)
}

func TestEncodeError_WithoutDetails(t *testing.T) {
	var buf bytes.Buffer
	err := EncodeError(&buf, 1, "internal_error", "something broke", nil)
	require.NoError(t, err)

	var decoded ErrorOutput
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, 1, decoded.ErrorCode)
	assert.Nil(t, decoded.Details)
}
