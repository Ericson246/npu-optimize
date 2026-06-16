package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Ericson246/npu-optimize/internal/output"
)

func fatal(exitCode int, errType string, msg string, args ...any) {
	slog.Error(msg, args...)
	_ = output.EncodeError(os.Stderr, exitCode, errType, msg, argsToMap(args))
	os.Exit(exitCode)
}

func argsToMap(args []any) map[string]any {
	if len(args) == 0 {
		return nil
	}
	m := make(map[string]any, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if ok {
			m[key] = fmt.Sprintf("%v", args[i+1])
		}
	}
	return m
}
