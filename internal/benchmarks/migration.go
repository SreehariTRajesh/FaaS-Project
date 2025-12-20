package benchmarks

import (
	"faas-migration/internal/container"
	"faas-migration/internal/cpu"
	"faas-migration/internal/kernel"
	"fmt"
	"log"
	"strconv"
	"syscall"
)

type BenchmarkRunner struct {
	// Add fields if necessary
	cgroupManager *container.CgroupManager
	clone3Exec    *container.Clone3Executor
	lm            *kernel.LatencyMetricsMonitor
	llc           *kernel.LLCMetricsMonitor
	frequencyMgr  *cpu.CPUFrequencyManager
}

func NewBenchmarkRunner(
	cgroupName string,
	latencyOutputFile string,
	cacheStatsOutputFile string,
) (*BenchmarkRunner, error) {

	cgroupManager, err := container.NewCgroupManager(cgroupName)

	if err != nil {
		return nil, fmt.Errorf("error while initializing cgroup manager %v", err)
	}

	clone3Exec, err := container.NewClone3Executor(cgroupManager.CgroupPath)

	if err != nil {
		return nil, fmt.Errorf("error while initializing clone3 executor %v", err)
	}

	latencyMetricsMonitor, err := kernel.NewLatencyMetricsMonitor(latencyOutputFile)

	if err != nil {
		return nil, fmt.Errorf("error while initializing latency metrics monitor %v", err)
	}

	cpuFreqManager, err := cpu.NewCPUFreqManager()

	if err != nil {
		log.Printf("warning: could not initialize CPU frequency manager: %v", err)
	}

	llcMetricsMonitor, err := kernel.NewLLCMetricsMonitor(cacheStatsOutputFile)

	if err != nil {
		return nil, fmt.Errorf("error while initializing LLC metrics monitor %v", err)
	}

	return &BenchmarkRunner{
		cgroupManager: cgroupManager,
		clone3Exec:    clone3Exec,
		lm:            latencyMetricsMonitor,
		llc:           llcMetricsMonitor,
		frequencyMgr:  cpuFreqManager,
	}, nil
}

func (br *BenchmarkRunner) RunBenchmark(pythonScript string, oldCPUSet string, memory string, newCPUSet string, oldCPUFreq uint64, newCPUFreq uint64) error {
	oldCPU, err := strconv.Atoi(oldCPUSet)
	if err != nil {
		return fmt.Errorf("invalid old CPU set: %w", err)
	}

	newCPU, err := strconv.Atoi(newCPUSet)
	if err != nil {
		return fmt.Errorf("invalid new CPU set: %w", err)
	}

	err = br.frequencyMgr.SetFrequency(oldCPU, oldCPUFreq)
	if err != nil {
		return fmt.Errorf("failed to set old CPU frequency: %w", err)
	}
	err = br.frequencyMgr.SetFrequency(newCPU, newCPUFreq)
	if err != nil {
		return fmt.Errorf("failed to set new CPU frequency: %w", err)
	}

	err = br.lm.Attach()
	if err != nil {
		return err
	}

	err = br.llc.AttachLLCEvents()
	if err != nil {
		return err
	}

	cgId := br.cgroupManager.GetCgroupId()
	doneChannel := make(chan struct{})

	go func() {
		err = br.lm.ReadEvents(doneChannel)
		if err != nil {
			log.Fatalf("error while reading events from ringbuffer: %v\n", err)
		}
	}()

	err = br.cgroupManager.SetCPUSet(oldCPUSet)
	if err != nil {
		return err
	}

	err = br.cgroupManager.SetMemory(memory)
	if err != nil {
		return err
	}

	err = br.lm.UpdateContainerCgroupId(cgId)
	if err != nil {
		return err
	}

	err = br.llc.UpdateContainerCgroupId(cgId)
	if err != nil {
		return err
	}

	pid := br.clone3Exec.CloneIntoCgroup(pythonScript)

	print("Process cloned into cgroup, CGroupID:", cgId, " PID:", pid, "\n")

	err = br.cgroupManager.SetCPUSet(newCPUSet)
	if err != nil {
		return err
	}

	<-doneChannel

	_ = syscall.Kill(int(pid), syscall.SIGKILL)

	if err != nil {
		return fmt.Errorf("failed to kill all processes in cgroup: %w", err)
	}

	br.clone3Exec.Close()
	br.cgroupManager.Close()
	br.lm.Close()

	return nil
}
