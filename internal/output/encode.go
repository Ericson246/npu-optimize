package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Ericson246/npu-optimize/internal/constants"
)

var schemaBaseURL = "https://Ericson246.github.io/npu-optimize/schemas"

func Encode(w io.Writer, o *Output) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(o)
}

func EncodeError(w io.Writer, code int, errType, msg string, details any) error {
	eo := ErrorOutput{
		Schema:    fmt.Sprintf("%s/error-v1.json", schemaBaseURL),
		Version:   1,
		Error:     true,
		ErrorCode: code,
		ErrorType: errType,
		Message:   msg,
		Details:   details,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(eo)
}

func New(schemaVersion int) *Output {
	return &Output{
		Schema:      fmt.Sprintf("%s/v%d.json", schemaBaseURL, schemaVersion),
		Version:     schemaVersion,
		GeneratedAt: time.Now().UTC(),
		ToolVersion: constants.Version,
		Backend:     "llama.cpp",
	}
}
