package recommend

import (
	"fmt"
	"strings"
	"time"
)

type ArchTier int

const (
	TierCuttingEdge ArchTier = iota + 1
	TierCurrentGen
	TierPreviousGen
	TierLegacy
	TierUnknown
)

func (t ArchTier) Score() float64 {
	switch t {
	case TierCuttingEdge:
		return 1.00
	case TierCurrentGen:
		return 0.85
	case TierPreviousGen:
		return 0.70
	case TierLegacy:
		return 0.50
	default:
		return 0.40
	}
}

func (t ArchTier) String() string {
	switch t {
	case TierCuttingEdge:
		return "cutting_edge"
	case TierCurrentGen:
		return "current_gen"
	case TierPreviousGen:
		return "previous_gen"
	case TierLegacy:
		return "legacy"
	default:
		return "unknown"
	}
}

type archRule struct {
	prefixes   []string
	tier       ArchTier
	minCreated time.Time // if zero, always matches
}

var archRules = []archRule{
	// Tier 1: Cutting edge — 2024/2025 architectures
	{prefixes: []string{"llama"}, tier: TierCuttingEdge,
		minCreated: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
	{prefixes: []string{"qwen2"}, tier: TierCuttingEdge,
		minCreated: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
	{prefixes: []string{"deepseek2"}, tier: TierCuttingEdge},
	{prefixes: []string{"gemma2"}, tier: TierCuttingEdge},
	{prefixes: []string{"gemma"}, tier: TierCuttingEdge,
		minCreated: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	{prefixes: []string{"phi3", "phi4"}, tier: TierCuttingEdge},
	{prefixes: []string{"command-r"}, tier: TierCuttingEdge},
	{prefixes: []string{"mistral-nemo"}, tier: TierCuttingEdge},
	{prefixes: []string{"dbrx"}, tier: TierCurrentGen},
	{prefixes: []string{"nemotron"}, tier: TierCuttingEdge},

	// Tier 2: Current generation
	{prefixes: []string{"mistral"}, tier: TierCurrentGen},
	{prefixes: []string{"mixtral"}, tier: TierCurrentGen},
	{prefixes: []string{"qwen2"}, tier: TierCurrentGen,
		minCreated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
	{prefixes: []string{"gemma"}, tier: TierCurrentGen,
		minCreated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
	{prefixes: []string{"yi"}, tier: TierCurrentGen},
	{prefixes: []string{"stablelm2"}, tier: TierCurrentGen},
	{prefixes: []string{"olmo"}, tier: TierCurrentGen},

	// Tier 3: Previous generation
	{prefixes: []string{"llama"}, tier: TierPreviousGen},
	{prefixes: []string{"qwen"}, tier: TierPreviousGen},
	{prefixes: []string{"falcon"}, tier: TierPreviousGen},
	{prefixes: []string{"phi2"}, tier: TierPreviousGen},
	{prefixes: []string{"starcoder"}, tier: TierPreviousGen},
	{prefixes: []string{"codegen"}, tier: TierPreviousGen},

	// Tier 4: Legacy
	{prefixes: []string{"gpt-neox"}, tier: TierLegacy},
	{prefixes: []string{"bloom"}, tier: TierLegacy},
	{prefixes: []string{"gptj"}, tier: TierLegacy},
	{prefixes: []string{"opt"}, tier: TierLegacy},
	{prefixes: []string{"mpt"}, tier: TierLegacy},
}

func classifyArch(arch string, createdAt time.Time) (ArchTier, float64) {
	lower := strings.ToLower(arch)

	for _, rule := range archRules {
		match := false
		for _, p := range rule.prefixes {
			if strings.HasPrefix(lower, p) {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		if !rule.minCreated.IsZero() && createdAt.Before(rule.minCreated) {
			continue
		}
		return rule.tier, rule.tier.Score()
	}

	// Fallback: score by age alone
	months := time.Since(createdAt).Hours() / (24 * 30)
	switch {
	case months < 6:
		return TierCuttingEdge, 0.80
	case months < 12:
		return TierCurrentGen, 0.70
	case months < 24:
		return TierPreviousGen, 0.55
	default:
		return TierUnknown, 0.40
	}
}

func estimateParamsFromName(modelID string) int64 {
	lower := strings.ToLower(modelID)
	idx := strings.Index(lower, "b")
	if idx < 0 {
		return 0
	}

	start := idx - 1
	for start >= 0 && (lower[start] >= '0' && lower[start] <= '9' || lower[start] == '.') {
		start--
	}
	start++
	if start >= idx {
		return 0
	}
	numStr := lower[start:idx]
	var val float64
	if _, err := fmt.Sscanf(numStr, "%f", &val); err != nil || val <= 0 {
		return 0
	}
	return int64(val * 1e9)
}
