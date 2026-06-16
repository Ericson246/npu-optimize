package recommend

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Ericson246/npu-optimize/internal/calculator"
	"github.com/Ericson246/npu-optimize/internal/hfclient"
	"github.com/Ericson246/npu-optimize/internal/hwinfo"
)

type Config struct {
	CtxSize           int
	VRAMMargin        int
	Mode              string
	AvailableMemoryMB int64
}

type Recommendation struct {
	Hardware         *hwinfo.Info           `json:"hardware"`
	Repo             string                 `json:"repo"`
	File             string                 `json:"file"`
	SHA256           string                 `json:"sha256,omitempty"`
	SizeBytes        int64                  `json:"size_bytes"`
	Architecture     string                 `json:"architecture"`
	ArchitectureType string                 `json:"architecture_type"`
	Multimodal       bool                   `json:"multimodal"`
	Header           *GGUFHeader            `json:"-"`
	FitsInVRAM       bool                   `json:"fits_in_vram"`
	VRAMResult       *calculator.VRAMResult `json:"vram_result"`
	Fallbacks        []Fallback             `json:"fallbacks,omitempty"`
}

type Fallback struct {
	File       string `json:"file"`
	SizeBytes  int64  `json:"size_bytes"`
	FitsInVRAM bool   `json:"fits_in_vram"`
	Reason     string `json:"reason"`
}

type Service struct {
	hfClient *hfclient.Client
	filter   FilterParams
	config   Config
}

func NewService(hf *hfclient.Client, config Config) *Service {
	return &Service{
		hfClient: hf,
		filter:   DefaultFilterParams(),
		config:   config,
	}
}

func (s *Service) Recommend(hw *hwinfo.Info) (*Recommendation, error) {
	models, err := s.searchModels()
	if err != nil {
		return nil, err
	}

	candidates := FilterModels(models, s.filter)
	if len(candidates) == 0 {
		return &Recommendation{Hardware: hw}, nil
	}

	memoryMB := s.config.AvailableMemoryMB
	if memoryMB <= 0 {
		if hw.GPU != nil {
			memoryMB = hw.GPU.VRAMFreeMB
		} else {
			memoryMB = 4000
		}
	}

	for _, candidate := range candidates {
		rec := s.tryRecommend(hw, candidate, memoryMB)
		if rec != nil {
			return rec, nil
		}
	}

	return &Recommendation{Hardware: hw}, nil
}

func (s *Service) tryRecommend(hw *hwinfo.Info, model hfclient.ModelInfo, memoryMB int64) *Recommendation {
	bestFile, bestSize := s.pickBestFile(model.Siblings)
	if bestFile == "" {
		return nil
	}

	if bestSize <= 0 {
		size, err := s.fetchFileSize(model.ID, bestFile)
		if err != nil {
			slog.Warn("cannot resolve file size, skipping", "repo", model.ID, "file", bestFile, "err", err)
			return nil
		}
		bestSize = size
	}

	headerData, err := s.hfClient.GetGGUFHeader(model.ID, bestFile)
	if err != nil {
		slog.Warn("cannot fetch GGUF header, skipping", "repo", model.ID, "file", bestFile, "err", err)
		return nil
	}

	header, err := ParseHeader(headerData)
	if err != nil {
		header = &GGUFHeader{NLayer: 28, NKVHeads: 4, NHeads: 32, HiddenSize: 2048, FileType: 10}
	}

	archType := "dense"
	if isMoE(header) {
		archType = "moe"
	}

	vramParams := calculator.Params{
		VRAMFreeMB: memoryMB,
		CtxSize:    s.config.CtxSize,
		VRAMMargin: s.config.VRAMMargin,
		FileSize:   bestSize,
		Header:     header,
	}
	vramResult := calculator.CalculateVRAM(vramParams)

	multimodal := false
	for _, t := range model.Tags {
		if t == "image-text-to-text" {
			multimodal = true
			break
		}
	}

	fallbacks := s.buildFallbacks(model.Siblings, bestFile, memoryMB, header)

	return &Recommendation{
		Hardware:         hw,
		Repo:             model.ID,
		File:             bestFile,
		SizeBytes:        bestSize,
		Architecture:     header.Architecture,
		ArchitectureType: archType,
		Multimodal:       multimodal,
		Header:           header,
		FitsInVRAM:       vramResult.FitsInVRAM,
		VRAMResult:       vramResult,
		Fallbacks:        fallbacks,
	}
}

func (s *Service) pickBestFile(siblings []hfclient.Sibling) (string, int64) {
	for _, sib := range siblings {
		if hasGGUFFile([]hfclient.Sibling{sib}, "Q4_K_M") {
			size := int64(0)
			if sib.Size != nil {
				size = *sib.Size
			}
			return sib.RFilename, size
		}
	}
	for _, sib := range siblings {
		if strings.HasSuffix(sib.RFilename, ".gguf") {
			size := int64(0)
			if sib.Size != nil {
				size = *sib.Size
			}
			return sib.RFilename, size
		}
	}
	return "", 0
}

func (s *Service) buildFallbacks(siblings []hfclient.Sibling, bestFile string, vramFreeMB int64, header *GGUFHeader) []Fallback {
	var fbs []Fallback

	for _, sib := range siblings {
		if !strings.HasSuffix(sib.RFilename, ".gguf") {
			continue
		}
		if sib.RFilename == bestFile {
			continue
		}
		if sib.Size == nil || *sib.Size <= 0 {
			continue
		}

		quant := extractQuant(sib.RFilename)
		if quant == "" {
			continue
		}

		vramParams := calculator.Params{
			VRAMFreeMB: vramFreeMB,
			CtxSize:    s.config.CtxSize,
			VRAMMargin: s.config.VRAMMargin,
			FileSize:   *sib.Size,
			Header:     header,
		}

		vramResult := calculator.CalculateVRAM(vramParams)

		if !vramResult.FitsInVRAM {
			continue
		}

		fbs = append(fbs, Fallback{
			File:       sib.RFilename,
			SizeBytes:  *sib.Size,
			FitsInVRAM: true,
			Reason:     "Alternativa con cuantización " + quant,
		})
	}

	sort.Slice(fbs, func(i, j int) bool {
		return fbs[i].SizeBytes > fbs[j].SizeBytes
	})

	if len(fbs) > 5 {
		fbs = fbs[:5]
	}

	return fbs
}

func (s *Service) searchModels() ([]hfclient.ModelInfo, error) {
	textModels, err := s.hfClient.SearchModels([]string{"gguf", "text-generation"}, 30)
	if err != nil {
		return nil, err
	}

	visionModels, err := s.hfClient.SearchModels([]string{"gguf", "image-text-to-text"}, 30)
	if err != nil {
		return textModels, nil
	}

	return mergeResults(textModels, visionModels), nil
}

func mergeResults(a, b []hfclient.ModelInfo) []hfclient.ModelInfo {
	seen := make(map[string]bool, len(a))
	merged := make([]hfclient.ModelInfo, 0, len(a)+len(b))

	for _, m := range a {
		seen[m.ModelID] = true
		merged = append(merged, m)
	}
	for _, m := range b {
		if !seen[m.ModelID] {
			merged = append(merged, m)
		}
	}
	return merged
}

func extractQuant(filename string) string {
	lower := strings.ToLower(filename)
	quants := []string{"q2_k", "q3_k_m", "q3_k_l", "q4_k_m", "q4_k_s", "q5_k_m", "q5_k_s", "q6_k", "q8_0", "f16"}
	for _, q := range quants {
		if strings.Contains(lower, q) {
			return strings.ToUpper(q)
		}
	}
	return ""
}

func (s *Service) fetchFileSize(repo, file string) (int64, error) {
	entries, err := s.hfClient.GetTree(repo)
	if err != nil {
		return 0, err
	}
	base := filepath.Base(file)
	for _, e := range entries {
		if e.LFS == nil {
			continue
		}
		if strings.EqualFold(e.Name, file) || strings.EqualFold(e.Name, base) {
			return e.LFS.Size, nil
		}
	}
	return 0, fmt.Errorf("file %s not found in tree of %s", file, repo)
}
