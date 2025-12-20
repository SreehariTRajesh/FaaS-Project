package cpu

import (
	"fmt"
	"path/filepath"
)

type Governor string

const (
	GovernorPerformance  Governor = "performance"
	GovernorPowersave    Governor = "powersave"
	GovernorUserspace    Governor = "userspace"
	GovernorOndemand     Governor = "ondemand"
	GovernorConservative Governor = "conservative"
	GovernorSchedutil    Governor = "schedutil"
)

func (m *CPUFrequencyManager) SetGovernor(cpuId int, governor Governor) error {
	if cpuId < 0 || cpuId >= m.numCPUS {
		return fmt.Errorf("invalid CPU Id: %d", cpuId)
	}

	cpuPath := filepath.Join(sysfsBasePath, fmt.Sprintf("cpu%d/cpufreq", cpuId))
	governorPath := filepath.Join(cpuPath, "scaling_governor")

	info, err := m.GetCPUInfo(cpuId)

	if err != nil {
		return fmt.Errorf("error while getting CPU info: %v", err)
	}
	// Check if the governor is available
	// for this CPU
	available := false
	for _, g := range info.AvailableGovernors {
		if g == string(governor) {
			available = true
			break
		}
	}

	if !available {
		return fmt.Errorf("governor %s not available for CPU %d", governor, cpuId)
	}

	return writeString(governorPath, string(governor))
}

func (m *CPUFrequencyManager) GetGovernor(cpuId int) (Governor, error) {
	info, err := m.GetCPUInfo(cpuId)

	if err != nil {
		return "", err
	}
	return Governor(info.Governor), nil
}
