package benchmarks

import (
	"faas-migration/internal/container"
	"faas-migration/internal/cpu"
	"faas-migration/internal/kernel"
	"fmt"
	"log"
	"strconv"
)

type FunctionBenchmarkRunner struct {
	// Add fields if necessary
	cgroupManager   *container.CgroupManager
	clone3Exec      *container.Clone3Executor
	frequencyMgr    *cpu.CPUFrequencyManager
	functionMonitor *kernel.FunctionMetricsMonitor
}

func NewFunctionBenchmarkRunner(
	cgroupName string,
	latencyOutputFile string) (*FunctionBenchmarkRunner, error) {

	cgroupManager, err := container.NewCgroupManager(cgroupName)

	if err != nil {
		return nil, fmt.Errorf("error while initializing cgroup manager %v", err)
	}

	clone3Exec, err := container.NewClone3Executor(cgroupManager.CgroupPath)

	if err != nil {
		return nil, fmt.Errorf("error while initializing clone3 executor %v", err)
	}

	cpuFreqManager, err := cpu.NewCPUFreqManager()

	if err != nil {
		log.Printf("warning: could not initialize CPU frequency manager: %v", err)
	}

	functionMonitor, err := kernel.NewFunctionMetricsMonitor(latencyOutputFile)
	if err != nil {
		return nil, fmt.Errorf("error while initializing proc runtime monitor: %v", err)
	}

	return &FunctionBenchmarkRunner{
		cgroupManager:   cgroupManager,
		clone3Exec:      clone3Exec,
		frequencyMgr:    cpuFreqManager,
		functionMonitor: functionMonitor,
	}, nil
}

func (f *FunctionBenchmarkRunner) RunFuncBenchmark(binary string, symbol, currCPUSet string, memory string, cpuFreq uint64) error {
	currCPU, err := strconv.Atoi(currCPUSet)
	if err != nil {
		return fmt.Errorf("invalid old CPU set: %w", err)
	}

	err = f.frequencyMgr.SetFrequency(currCPU, cpuFreq)
	if err != nil {
		return fmt.Errorf("failed to set old CPU frequency: %w", err)
	}

	err = f.functionMonitor.Attach(&binary, &symbol)
	if err != nil {
		return fmt.Errorf("error while attaching proc monitor events: %w", err)
	}

	f.clone3Exec.CloneIntoCgroup(binary)

	doneChannel := make(chan struct{})
	err = f.functionMonitor.ReadEvents(doneChannel)

	if err != nil {
		return fmt.Errorf("error while reading function events: %v", err)
	}

	return nil
}
