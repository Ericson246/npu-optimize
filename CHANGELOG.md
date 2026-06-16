# Changelog

## [0.1.1] - 2026-06-16

### Fixed
- Model selection uses best-fit instead of first-fit: now selects the largest
  model that fits in VRAM instead of the first popular one (#1)
- Batch file size resolution via HF paths-info API (more efficient than GetTree)
- Increased candidate pool from 8 to 30 for better coverage

## [0.1.0] - 2026-06-15

### Added
- `detect` command: hardware detection + model recommendation
- HuggingFace API integration (search, tree, GGUF headers)
- GGUF parser for model metadata extraction
- VRAM calculator with ctx_max estimation
- Cache system for hardware fingerprints
- JSON output with versioned schema and error responses
- Hardware detection: NVIDIA (nvidia-smi), Intel iGPU (vulkaninfo), CPU fallback
- Support Matrix: exit codes 0-4, auth detection, error output contract
- README
- Full CI/CD: lint + test + build (Windows + Linux), goreleaser publishing
