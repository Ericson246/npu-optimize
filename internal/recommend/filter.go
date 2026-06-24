package recommend

import (
	"time"

	"github.com/Ericson246/npu-optimize/internal/hfclient"
)

const maxModelAgeMonths = 6

func FilterByAge(models []hfclient.ModelInfo) []hfclient.ModelInfo {
	cutoff := time.Now().AddDate(0, -maxModelAgeMonths, 0)
	var result []hfclient.ModelInfo
	for _, m := range models {
		if !m.CreatedAt.IsZero() && m.CreatedAt.Before(cutoff) {
			continue
		}
		result = append(result, m)
	}
	return result
}

func FilterByPipelineTag(models []hfclient.ModelInfo) []hfclient.ModelInfo {
	var result []hfclient.ModelInfo
	for _, m := range models {
		switch m.PipelineTag {
		case "text-generation", "image-text-to-text":
			result = append(result, m)
		}
	}
	return result
}
