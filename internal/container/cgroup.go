package container

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type CgroupManager struct {
	CgroupPath string
	CgroupId   uint32
	CgroupName string
}

func NewCgroupManager(cgroupName string) (*CgroupManager, error) {
	cgroupPath := filepath.Join("/sys/fs/cgroup", cgroupName)

	err := os.Mkdir(cgroupPath, 0755)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("failed to create cgroup directory %s: %w", cgroupPath, err)
	}

	info, err := os.Stat(cgroupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat cgroup directory %s: %w", cgroupPath, err)
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("failed to get stat info for cgroup directory %s", cgroupPath)
	}

	return &CgroupManager{
		CgroupPath: cgroupPath,
		CgroupName: cgroupName,
		CgroupId:   uint32(stat.Ino),
	}, nil
}

func (cm *CgroupManager) SetCPUSet(cpuSet string) error {
	cpusetPath := filepath.Join(cm.CgroupPath, "cpuset.cpus")
	fmt.Println(cpusetPath)
	err := os.WriteFile(cpusetPath, []byte(cpuSet), 0644)
	if err != nil {
		fmt.Printf("Failed to write to cpuset file %s: %v\n", cpusetPath, err)
		return err
	}

	fmt.Println("Successfully set cpuset to: ", cpuSet)

	return nil
}

func (cm *CgroupManager) SetMemory(memoryMB string) error {
	memoryPath := filepath.Join(cm.CgroupPath, "memory.max")
	err := os.WriteFile(memoryPath, []byte(memoryMB), 0644)
	if err != nil {
		fmt.Printf("Failed to write to memory file %s: %v\n", memoryPath, err)
		return err
	}
	return nil
}

func (cm *CgroupManager) Close() error {
	err := os.RemoveAll(cm.CgroupPath)
	if err != nil {
		return fmt.Errorf("failed to delete cgroup directory %s: %w", cm.CgroupPath, err)
	}
	return nil
}

func (cm *CgroupManager) GetCgroupId() uint32 {
	return cm.CgroupId
}

func (cm *CgroupManager) KillAllInCgroup() error {
	procsPath := filepath.Join(cm.CgroupPath, "cgroup.procs")

	for attempt := 1; attempt <= 3; attempt++ {
		data, err := os.ReadFile(procsPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("failed to read cgroup.procs: %w", err)
		}

		pids := strings.Fields(string(data))
		if len(pids) == 0 {
			return nil
		}

		for _, pidStr := range pids {
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				continue
			}

			p, err := os.FindProcess(pid)

			if err == nil {
				_ = p.Signal(syscall.SIGKILL)
			}

			_, _ = p.Wait()
			fmt.Printf("Killed process with PID: %d\n", pid)
		}
	}

	data, _ := os.ReadFile(procsPath)
	if len(strings.Fields(string(data))) > 0 {
		return fmt.Errorf("failed to kill all processes; cgroup is not empty")
	}

	return nil
}
