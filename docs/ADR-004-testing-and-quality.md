# ADR-004: Testing and Quality

**Date:** 2026-06-14
**Status:** Accepted
**Repo:** `github.com/Ericson246/npu-optimize`

## Context

`npu-optimize` interacts with multiple external systems (real hardware, HuggingFace API, GitHub Releases, llama-bench as subprocess). This requires a clear testing strategy: fast deterministic unit tests with mocked data, opt-in integration tests against real systems, and minimal manual E2E verification.

As an open-source project with free CI (GitHub Actions), tests must be fast (< 2 min) and not depend on specialized hardware.

## Decision

### Tools

| Tool | Purpose | Discarded alternative |
|:-----|:--------|:----------------------|
| `testing` (stdlib) | Standard Go test framework | — |
| `testify` | Assertions (`assert.Equal`, `require.NoError`) | Plain `testing` (more boilerplate) |
| `golangci-lint` | Static linting | `go vet` alone (less coverage) |
| `go test -race` | Data race detection (Linux only; Windows requires CGO) | — |
| `go test -coverprofile` | Code coverage (planned for CI) | — |
| `go test -fuzz` | Fuzzing parsers (planned for v0.2.0) | — |
| `httptest` (stdlib) | HTTP mock for HF API and GitHub API | — |

`testify` is used **only for assertions** (`assert`/`require`). The `mock` package from testify is not used: mocks are built with native Go interfaces (ADR-001) and `httptest.Server` for HTTP. This keeps dependencies minimal and tests more readable.

### Test Pyramid

```
         ╱──────╲
        ╱  E2E   ╲           ← Manual: real benchmark with GPU
       ╱──────────╲
      ╱ Integration ╲         ← `go test -tags=integration`
     ╱              ╲         ← Opt-in, requires HF token or llamabench
    ╱────────────────╲
   ╱──────────────────╲
  ╱   Unit tests         ╲     ← `go test ./...` (always in CI)
 ╱                        ╲   ← 90%+ of what's tested
╱──────────────────────────╲
```

### Test Types by Package

| Package | Unit | Integration | Mock strategy |
|:--------|:-----|:------------|:--------------|
| `hwinfo` | ✅ | — | `DetectorFunc` injectable interface. GPU name/VRAM tests parse inline strings |
| `hfclient` | ✅ | ✅ opt-in (planned) | `httptest.Server` with inline response builders |
| `calculator` | ✅ | — | Pure: formulas only. External test package (`calculator_test`) |
| `recommend/filter` | ✅ | — | Inline model structs, no external data |
| `recommend/gguf` | ✅ | — | `buildGGUF()` helper generates synthetic headers inline |
| `recommend/recommend` | ✅ | — | Mock hfclient via exported `BaseURL`/`HTTPClient` fields + `hwinfo.Info` direct |
| `cache` | ✅ | — | `os.TempDir`, no mocks |
| `constants` | ✅ | — | Inline assertions on constant values |
| `output` | ✅ | — | `bytes.Buffer` for JSON encode/decode roundtrip |
| `cmd/` | ✅ | — | Native Go tests on `resolveDetectConfig`, `getToken` |
| `backend/backend` | — | — | Pure interface, no tests yet |
| `backend/llamacpp` | — | — | Stub, tested in v0.2.0 |
| `logger` | — | — | No tests yet |

### testdata

No `testdata/` directory in v0.1.0. All tests use inline data (string literals, `buildGGUF()`, `httptest.Server`).
Planned for v0.2.0 when benchmark pipeline is implemented.

### Fuzz Testing

Not implemented yet — planned for v0.2.0 (targets: GGUF parser, HF API responses).

### Coverage Targets

Coverage tracking is not yet set up in CI. Planned for v0.2.0.
Current status: informational only (run locally with `go test -coverprofile`).

### CI/CD (GitHub Actions)

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - run: go test ./... -v -count=1

  test-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - run: go test ./... -v -count=1

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - run: go build ./cmd/npu-optimize
```

Integration tests (`//go:build integration`) are planned for v0.2.0.

### Integration Tests

Not implemented yet — planned for v0.2.0 (real HF API calls, real GGUF header parsing).
Marked with `//go:build integration` build tag and run via:

```bash
go test -tags=integration ./...
```

### Manual Tests (E2E)

Not automated in CI due to GPU requirement:

1. `npu-optimize detect` on hardware with NVIDIA GPU
2. `npu-optimize detect` on CPU-only hardware
3. `npu-optimize detect --mode cpu` with insufficient RAM (verify exit code 3)

### Best Practices

- **Table-driven tests**: standard Go pattern with `[]struct{...}` for multiple cases
- **t.Parallel()**: independent tests marked as parallel
- **t.Cleanup()**: temporary resource cleanup
- **Subtests**: `t.Run("name", ...)` for grouping related cases
- **No testmain**: global Setup/Teardown not needed

## References

- ADR-001: Interfaces (`backend.Interface`, `hwinfo.Detector`, etc.) enabling mocks
- ADR-003: Exit codes and output schema (validation in tests)
- `github.com/stretchr/testify`: assertions
- `github.com/golangci/golangci-lint`: linter
- Go blog: "Table Driven Tests" (https://go.dev/blog/subtests)
