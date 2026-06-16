# Changelog

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
