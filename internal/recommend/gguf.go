package recommend

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

const ggufMagic = 0x46554747

type GGUFHeader struct {
	NLayer        int    `json:"n_layers"`
	NKVHeads      int    `json:"n_kv_heads"`
	NHeads        int    `json:"n_heads"`
	HiddenSize    int    `json:"hidden_size"`
	FileType      int    `json:"file_type"`
	NExperts      *int   `json:"n_experts,omitempty"`
	NExpertsUsed  *int   `json:"n_experts_used,omitempty"`
	ExpertFFNSize *int   `json:"expert_ffn_size,omitempty"`
	FFNSize       *int   `json:"ffn_size,omitempty"`
	NMTPHeads     *int   `json:"n_mtp_heads,omitempty"`
	VisionDim     *int   `json:"vision_dim,omitempty"`
	VisionLayers  *int   `json:"vision_layers,omitempty"`
	Architecture  string `json:"architecture"`
	NumParameters int64  `json:"num_parameters,omitempty"`
}

func ParseHeader(data []byte) (*GGUFHeader, error) {
	r := reader{data: data}

	magic := r.readU32()
	if magic != ggufMagic {
		return nil, fmt.Errorf("invalid GGUF magic: %x", magic)
	}

	_ = r.readU32()
	_ = r.readU64()
	metadataKvCount := r.readU64()

	h := &GGUFHeader{}

	for i := uint64(0); i < metadataKvCount; i++ {
		key := r.readString()
		value := r.readValue()

		switch {
		case strings.HasSuffix(key, ".vision.block_count"):
			if n := toInt(value); n > 0 {
				h.VisionLayers = &n
			}
		case strings.HasSuffix(key, ".vision.embedding_length"):
			if n := toInt(value); n > 0 {
				h.VisionDim = &n
			}
		case strings.HasSuffix(key, ".block_count"):
			h.NLayer = toInt(value)
		case strings.HasSuffix(key, ".attention.head_count_kv"):
			h.NKVHeads = toInt(value)
		case strings.HasSuffix(key, ".attention.head_count"):
			h.NHeads = toInt(value)
		case strings.HasSuffix(key, ".attention.hidden_size"):
			h.HiddenSize = toInt(value)
		case key == "general.file_type":
			h.FileType = toInt(value)
		case strings.HasSuffix(key, ".expert_used_count"):
			if n := toInt(value); n > 0 {
				h.NExpertsUsed = &n
			}
		case strings.HasSuffix(key, ".expert_count"):
			if n := toInt(value); n > 0 {
				h.NExperts = &n
			}
		case strings.HasSuffix(key, ".expert_feed_forward_length"):
			if n := toInt(value); n > 0 {
				h.ExpertFFNSize = &n
			}
		case strings.HasSuffix(key, ".feed_forward_length"):
			if n := toInt(value); n > 0 {
				h.FFNSize = &n
			}
		case strings.HasSuffix(key, ".num_nextn_predict"):
			if n := toInt(value); n > 0 {
				h.NMTPHeads = &n
			}
		case key == "general.architecture":
			h.Architecture = toString(value)
		case key == "general.parameter_count":
			h.NumParameters = toInt64(value)
		}
	}

	if r.err != nil {
		return nil, r.err
	}
	return h, nil
}

func toInt(v any) int {
	switch val := v.(type) {
	case uint64:
		return int(val)
	case int64:
		return int(val)
	case float64:
		return int(val)
	case bool:
		if val {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func toInt64(v any) int64 {
	switch val := v.(type) {
	case uint64:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	default:
		return 0
	}
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

type reader struct {
	data []byte
	pos  int
	err  error
}

func (r *reader) readU32() uint32 {
	if r.err != nil {
		return 0
	}
	if r.pos+4 > len(r.data) {
		r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
		return 0
	}
	v := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return v
}

func (r *reader) readU64() uint64 {
	if r.err != nil {
		return 0
	}
	if r.pos+8 > len(r.data) {
		r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
		return 0
	}
	v := binary.LittleEndian.Uint64(r.data[r.pos:])
	r.pos += 8
	return v
}

func (r *reader) readString() string {
	if r.err != nil {
		return ""
	}
	length := r.readU64()
	if r.err != nil {
		return ""
	}
	if int(length) > len(r.data)-r.pos {
		r.err = fmt.Errorf("gguf: string length %d exceeds remaining data at offset %d", length, r.pos)
		return ""
	}
	s := string(r.data[r.pos : r.pos+int(length)])
	r.pos += int(length)
	return s
}

func (r *reader) readValue() any {
	if r.err != nil {
		return nil
	}
	valueType := r.readU32()
	if r.err != nil {
		return nil
	}

	switch valueType {
	case 0:
		if r.pos+1 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := r.data[r.pos]
		r.pos++
		return uint64(v)
	case 1:
		if r.pos+1 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := int8(r.data[r.pos])
		r.pos++
		return int64(v)
	case 2:
		if r.pos+2 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := binary.LittleEndian.Uint16(r.data[r.pos:])
		r.pos += 2
		return uint64(v)
	case 3:
		if r.pos+2 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := int16(binary.LittleEndian.Uint16(r.data[r.pos:]))
		r.pos += 2
		return int64(v)
	case 4:
		if r.pos+4 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := binary.LittleEndian.Uint32(r.data[r.pos:])
		r.pos += 4
		return uint64(v)
	case 5:
		if r.pos+4 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := int32(binary.LittleEndian.Uint32(r.data[r.pos:]))
		r.pos += 4
		return int64(v)
	case 6:
		if r.pos+4 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := math.Float32frombits(binary.LittleEndian.Uint32(r.data[r.pos:]))
		r.pos += 4
		return float64(v)
	case 7:
		if r.pos+1 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := r.data[r.pos]
		r.pos++
		return v != 0
	case 8:
		return r.readString()
	case 9:
		elemType := r.readU32()
		count := r.readU64()
		for i := uint64(0); i < count; i++ {
			r.skipValue(elemType)
		}
		return nil
	case 10:
		return r.readU64()
	case 11:
		if r.pos+8 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := int64(binary.LittleEndian.Uint64(r.data[r.pos:]))
		r.pos += 8
		return v
	case 12:
		if r.pos+8 > len(r.data) {
			r.err = fmt.Errorf("gguf: unexpected end of data at offset %d", r.pos)
			return nil
		}
		v := math.Float64frombits(binary.LittleEndian.Uint64(r.data[r.pos:]))
		r.pos += 8
		return v
	default:
		return nil
	}
}

func (r *reader) skipValue(vtype uint32) {
	if r.err != nil {
		return
	}
	switch vtype {
	case 0, 1, 7:
		r.pos++
	case 2, 3:
		r.pos += 2
	case 4, 5, 6:
		r.pos += 4
	case 8:
		r.pos += int(r.readU64())
	case 9:
		elemType := r.readU32()
		count := r.readU64()
		for i := uint64(0); i < count; i++ {
			r.skipValue(elemType)
		}
	case 10, 11, 12:
		r.pos += 8
	}
}

func (h *GGUFHeader) NumLayer() int      { return h.NLayer }
func (h *GGUFHeader) NumKVHeads() int    { return h.NKVHeads }
func (h *GGUFHeader) HeadDim() int       { return HeadDim(h) }
func (h *GGUFHeader) QuantFileType() int { return h.FileType }

func isMoE(h *GGUFHeader) bool {
	return h.NExperts != nil && *h.NExperts > 0
}

func HeadDim(h *GGUFHeader) int {
	if h.HiddenSize > 0 && h.NHeads > 0 {
		return h.HiddenSize / h.NHeads
	}
	return 128
}
