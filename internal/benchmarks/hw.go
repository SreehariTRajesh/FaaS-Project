package benchmarks

import (
	"faas-migration/internal/kernel"
	"fmt"
	"log"
)

type HardwareBenchmarkRunner struct {
	hw *kernel.HardwareMetricsMonitor
}

func NewHardwarBenchmarkRunner(outputFilePath string) (*HardwareBenchmarkRunner, error) {
	hw, err := kernel.NewHardwareMetricsMonitor(outputFilePath)
	if err != nil {
		return nil, fmt.Errorf("error while initializing hardware metrics monitor: %w", err)
	}
	return &HardwareBenchmarkRunner{
		hw: hw,
	}, nil
}

func (h *HardwareBenchmarkRunner) RunBenchmark() error {
	err := h.hw.AttachHardwareEvents()
	if err != nil {
		return fmt.Errorf("error while attaching hardware events: %w", err)
	}
	log.Println("successfully attached hardware perf events")

	err = h.hw.ReadStats(10)

	if err != nil {
		return fmt.Errorf("error while reading stats: %w", err)
	}
	return nil
}
