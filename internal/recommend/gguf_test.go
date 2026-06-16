package recommend

import (
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildGGUF(kv map[string]any) []byte {
	var buf []byte
	put := func(v ...byte) { buf = append(buf, v...) }

	putU32 := func(v uint32) {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, v)
		put(b...)
	}
	putU64 := func(v uint64) {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, v)
		put(b...)
	}

	putU32(ggufMagic)
	putU32(3)
	putU64(0)
	putU64(uint64(len(kv)))

	for key, val := range kv {
		putU64(uint64(len(key)))
		put([]byte(key)...)

		switch v := val.(type) {
		case uint64:
			putU32(10)
			putU64(v)
		case uint32:
			putU32(4)
			putU32(v)
		case string:
			putU32(8)
			putU64(uint64(len(v)))
			put([]byte(v)...)
		case bool:
			putU32(7)
			if v {
				put(1)
			} else {
				put(0)
			}
		case float64:
			putU32(12)
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, math.Float64bits(v))
			put(b...)
		default:
			panic(fmt.Sprintf("unsupported type %T", val))
		}
	}

	return buf
}

func TestParseHeader_Valid(t *testing.T) {
	data := buildGGUF(map[string]any{
		"general.architecture":             "llama",
		"llama.block_count":                uint32(32),
		"llama.attention.head_count_kv":    uint32(8),
		"llama.attention.head_count":       uint32(32),
		"llama.attention.hidden_size":      uint32(4096),
		"general.file_type":                uint32(10),
		"llama.expert_count":               uint32(8),
		"llama.expert_used_count":          uint32(2),
		"llama.num_nextn_predict":          uint32(3),
		"llama.vision.embedding_length":    uint32(1024),
		"llama.vision.block_count":         uint32(24),
		"llama.expert_feed_forward_length": uint32(1024),
		"llama.feed_forward_length":        uint32(4096),
	})

	header, err := ParseHeader(data)
	require.NoError(t, err)
	assert.Equal(t, "llama", header.Architecture)
	assert.Equal(t, 32, header.NLayer)
	assert.Equal(t, 8, header.NKVHeads)
	assert.Equal(t, 32, header.NHeads)
	assert.Equal(t, 4096, header.HiddenSize)
	assert.Equal(t, 10, header.FileType)
	require.NotNil(t, header.NExperts)
	assert.Equal(t, 8, *header.NExperts)
	require.NotNil(t, header.NExpertsUsed)
	assert.Equal(t, 2, *header.NExpertsUsed)
	require.NotNil(t, header.NMTPHeads)
	assert.Equal(t, 3, *header.NMTPHeads)
	require.NotNil(t, header.VisionDim)
	assert.Equal(t, 1024, *header.VisionDim)
	require.NotNil(t, header.VisionLayers)
	assert.Equal(t, 24, *header.VisionLayers)
	require.NotNil(t, header.ExpertFFNSize)
	assert.Equal(t, 1024, *header.ExpertFFNSize)
	require.NotNil(t, header.FFNSize)
	assert.Equal(t, 4096, *header.FFNSize)
}

func TestParseHeader_InvalidMagic(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00}
	_, err := ParseHeader(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GGUF magic")
}

func TestParseHeader_AllZeros(t *testing.T) {
	_, err := ParseHeader([]byte{0x00, 0x00, 0x00, 0x00})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GGUF magic")
}

func TestParseHeader_Minimal(t *testing.T) {
	data := buildGGUF(map[string]any{
		"general.architecture": "qwen2",
		"llama.block_count":    uint32(28),
	})

	header, err := ParseHeader(data)
	require.NoError(t, err)
	assert.Equal(t, "qwen2", header.Architecture)
	assert.Equal(t, 28, header.NLayer)
	assert.Equal(t, 0, header.NKVHeads)
}

func TestParseHeader_EmptyKV(t *testing.T) {
	data := buildGGUF(nil)
	header, err := ParseHeader(data)
	require.NoError(t, err)
	assert.Empty(t, header.Architecture)
}

func TestIsMoE(t *testing.T) {
	assert.False(t, isMoE(&GGUFHeader{}))

	experts := 8
	assert.True(t, isMoE(&GGUFHeader{NExperts: &experts}))

	zero := 0
	assert.False(t, isMoE(&GGUFHeader{NExperts: &zero}))
}

func TestHeadDim(t *testing.T) {
	assert.Equal(t, 128, HeadDim(&GGUFHeader{}))

	assert.Equal(t, 128, HeadDim(&GGUFHeader{HiddenSize: 4096, NHeads: 32}))
	assert.Equal(t, 64, HeadDim(&GGUFHeader{HiddenSize: 4096, NHeads: 64}))
	assert.Equal(t, 128, HeadDim(&GGUFHeader{HiddenSize: 0, NHeads: 32}))
	assert.Equal(t, 128, HeadDim(&GGUFHeader{HiddenSize: 4096, NHeads: 0}))
}

func TestReader_ReadString(t *testing.T) {
	data := make([]byte, 8+5)
	binary.LittleEndian.PutUint64(data, 5)
	copy(data[8:], "hello")
	r := &reader{data: data}
	assert.Equal(t, "hello", r.readString())
}

func TestReader_ReadStringTruncated(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 100)
	r := &reader{data: data}
	assert.Empty(t, r.readString())
}

func TestReader_ReadU32(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 42)
	r := &reader{data: data}
	assert.Equal(t, uint32(42), r.readU32())
}

func TestReader_ReadU32Truncated(t *testing.T) {
	r := &reader{data: []byte{0, 0}}
	assert.Equal(t, uint32(0), r.readU32())
}

func TestReader_ReadU64(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, 42)
	r := &reader{data: data}
	assert.Equal(t, uint64(42), r.readU64())
}

func TestReader_ReadU64Truncated(t *testing.T) {
	r := &reader{data: []byte{0, 0}}
	assert.Equal(t, uint64(0), r.readU64())
}

func TestToInt(t *testing.T) {
	assert.Equal(t, 42, toInt(uint64(42)))
	assert.Equal(t, -1, toInt(int64(-1)))
	assert.Equal(t, 3, toInt(float64(3.14)))
	assert.Equal(t, 1, toInt(true))
	assert.Equal(t, 0, toInt(false))
	assert.Equal(t, 0, toInt("string"))
	assert.Equal(t, 0, toInt(nil))
}

func TestToString(t *testing.T) {
	assert.Equal(t, "hello", toString("hello"))
	assert.Equal(t, "world", toString([]byte("world")))
	assert.Equal(t, "42", toString(42))
}
