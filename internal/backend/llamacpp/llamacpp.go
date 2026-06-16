package llamacpp

import (
	"errors"

	"github.com/Ericson246/npu-optimize/internal/backend"
	"github.com/Ericson246/npu-optimize/internal/hwinfo"
)

type Backend struct{}

func New() *Backend { return &Backend{} }

func (b *Backend) Type() backend.Type { return backend.TypeLlamaCpp }

func (b *Backend) Detect(hw *hwinfo.Info) bool {
	return hw.GPU != nil
}

func (b *Backend) Fit(modelPath string) (*backend.FitResult, error) {
	return nil, errors.New("not implemented until v0.2.0")
}

func (b *Backend) Benchmark(modelPath string, p backend.Params) (*backend.BenchResult, error) {
	return nil, errors.New("not implemented until v0.2.0")
}

func (b *Backend) Sweep(modelPath string, baseline backend.Params, mode string) ([]backend.BenchResult, error) {
	return nil, errors.New("not implemented until v0.2.0")
}
