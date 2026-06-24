package recommend

import (
	"fmt"
	"strings"

	"github.com/Ericson246/npu-optimize/internal/hfclient"
)

var ggufRangeSizes = []int{512 * 1024, 1024 * 1024, 2 * 1024 * 1024, 4 * 1024 * 1024, 8 * 1024 * 1024, 16 * 1024 * 1024}

func fetchAndParseHeader(client *hfclient.Client, repo, file string) (*GGUFHeader, error) {
	var lastErr error
	for _, size := range ggufRangeSizes {
		data, err := client.GetGGUFHeader(repo, file, size)
		if err != nil {
			lastErr = err
			continue
		}
		header, err := ParseHeader(data)
		if err == nil {
			return header, nil
		}
		lastErr = err
		if !strings.Contains(err.Error(), "unexpected end of data") {
			return nil, err
		}
	}
	return nil, fmt.Errorf("cannot parse GGUF header after %d attempts: %w", len(ggufRangeSizes), lastErr)
}
