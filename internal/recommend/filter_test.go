package recommend

import (
	"testing"
	"time"

	"github.com/Ericson246/npu-optimize/internal/hfclient"
	"github.com/stretchr/testify/assert"
)

func TestFilterByAge_KeepsRecent(t *testing.T) {
	now := time.Now()
	recent := now.AddDate(0, -1, 0)

	models := []hfclient.ModelInfo{
		{ModelID: "recent-model", CreatedAt: recent},
		{ModelID: "old-model", CreatedAt: now.AddDate(0, -7, 0)},
	}

	result := FilterByAge(models)
	assert.Len(t, result, 1)
	assert.Equal(t, "recent-model", result[0].ModelID)
}

func TestFilterByAge_AllRecent(t *testing.T) {
	models := []hfclient.ModelInfo{
		{ModelID: "a", CreatedAt: time.Now()},
		{ModelID: "b", CreatedAt: time.Now().AddDate(0, -2, 0)},
	}
	result := FilterByAge(models)
	assert.Len(t, result, 2)
}

func TestFilterByAge_Empty(t *testing.T) {
	result := FilterByAge(nil)
	assert.Empty(t, result)
}

func TestFilterByAge_ZeroTime(t *testing.T) {
	models := []hfclient.ModelInfo{
		{ModelID: "no-date"},
	}
	result := FilterByAge(models)
	assert.Len(t, result, 1)
}

func TestFilterByPipelineTag_KeepsTextGeneration(t *testing.T) {
	models := []hfclient.ModelInfo{
		{ModelID: "llm", PipelineTag: "text-generation"},
		{ModelID: "vlm", PipelineTag: "image-text-to-text"},
		{ModelID: "image", PipelineTag: "text-to-image"},
		{ModelID: "embed", PipelineTag: "feature-extraction"},
	}
	result := FilterByPipelineTag(models)
	assert.Len(t, result, 2)
	assert.Equal(t, "llm", result[0].ModelID)
	assert.Equal(t, "vlm", result[1].ModelID)
}

func TestFilterByPipelineTag_Empty(t *testing.T) {
	assert.Empty(t, FilterByPipelineTag(nil))
	assert.Empty(t, FilterByPipelineTag([]hfclient.ModelInfo{}))
}

func TestFilterByPipelineTag_NoMatch(t *testing.T) {
	models := []hfclient.ModelInfo{
		{ModelID: "img-gen", PipelineTag: "text-to-image"},
	}
	result := FilterByPipelineTag(models)
	assert.Empty(t, result)
}
