package benchmarks

import (
	"faas-migration/internal/container"
	"faas-migration/internal/cpu"
	"faas-migration/internal/kernel"
	"fmt"
	"log"
	"strconv"
)

type ProcBenchmarkRunner struct {
	// Add fields if necessary
	cgroupManager *container.CgroupManager
	clone3Exec    *container.Clone3Executor
	frequencyMgr  *cpu.CPUFrequencyManager
	procMonitor   *kernel.ProcRuntimeMonitor
	hwMonitor     *kernel.HardwareMetricsMonitor
}

func NewProcBenchmarkRunner(
	cgroupName string,
	procOutputFile string,
	hardwareOutputFile string,
) (*ProcBenchmarkRunner, error) {

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

	procMonitor, err := kernel.NewProcRuntimeMonitor(procOutputFile)
	if err != nil {
		return nil, fmt.Errorf("error while initializing proc runtime monitor: %v", err)
	}

	hwMonitor, err := kernel.NewHardwareMetricsMonitor(hardwareOutputFile)
	if err != nil {
		return nil, fmt.Errorf("error while initializing hardware metrics monitor: %v", err)
	}

	return &ProcBenchmarkRunner{
		cgroupManager: cgroupManager,
		clone3Exec:    clone3Exec,
		frequencyMgr:  cpuFreqManager,
		procMonitor:   procMonitor,
		hwMonitor:     hwMonitor,
	}, nil
}

func (p *ProcBenchmarkRunner) RunProcBenchmark(pythonScript string, currCPUSet string, memory string, cpuFreq uint64) error {
	currCPU, err := strconv.Atoi(currCPUSet)
	if err != nil {
		return fmt.Errorf("invalid old CPU set: %w", err)
	}

	err = p.frequencyMgr.SetFrequency(currCPU, cpuFreq)
	if err != nil {
		return fmt.Errorf("failed to set old CPU frequency: %w", err)
	}

	err = p.procMonitor.Attach()
	if err != nil {
		return fmt.Errorf("error while attaching proc monitor events: %w", err)
	}

	err = p.hwMonitor.AttachHardwareEvents()
	if err != nil {
		return fmt.Errorf("error while attaching hardware monitor events: %w", err)
	}

	cgId := p.cgroupManager.GetCgroupId()
	//doneChannel := make(chan struct{})

	err = p.procMonitor.UpdateContainerCgroupId(cgId)
	if err != nil {
		return err
	}

	err = p.hwMonitor.UpdateContainerCgroupId(cgId)
	if err != nil {
		return err
	}

	doneChannel := make(chan struct{})

	p.clone3Exec.CloneIntoCgroup(pythonScript)

	err = p.procMonitor.ReadEvents(doneChannel)
	if err != nil {
		return fmt.Errorf("error while reading events from ringbuffer: %v\n", err)
	}

	return nil
}
