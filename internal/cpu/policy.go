package cpu

import (
	"fmt"
	"path/filepath"
	"strconv"
)

func (m *CPUFrequencyManager) SetMinFrequency(cpuId int, freqKHz uint64) error {
	if cpuId < 0 || cpuId >= m.numCPUS {
		return fmt.Errorf("invalid CPU ID: %d", cpuId)
	}

	info, err := m.GetCPUInfo(cpuId)
	if err != nil {
		return err
	}

	if freqKHz < info.MinFreq || freqKHz > info.MaxFreq {
		return fmt.Errorf("frequency %d kHz out of range [%d, %d]", freqKHz, info.MinFreq, info.MaxFreq)
	}

	cpuPath := filepath.Join(sysfsBasePath, fmt.Sprintf("cpu%d/cpufreq", cpuId))
	minFreqPath := filepath.Join(cpuPath, "scaling_min_freq")

	return writeString(minFreqPath, strconv.FormatUint(freqKHz, 10))
}

func (m *CPUFrequencyManager) SetMaxFrequency(cpuId int, freqKHz uint64) error {
	if cpuId < 0 || cpuId >= m.numCPUS {
		return fmt.Errorf("invalid CPU Id: %d", cpuId)
	}

	info, err := m.GetCPUInfo(cpuId)

	if err != nil {
		return err
	}

	if freqKHz < info.MinFreq || freqKHz > info.MaxFreq {
		return fmt.Errorf("frequency %d kHz out of range [%d, %d]", freqKHz, info.MinFreq, info.MaxFreq)
	}

	cpuPath := filepath.Join(sysfsBasePath, fmt.Sprintf("cpu%d/cpufreq", cpuId))
	maxFreqPath := filepath.Join(cpuPath, "scaling_max_freq")

	return writeString(maxFreqPath, strconv.FormatUint(freqKHz, 10))
}

func (m *CPUFrequencyManager) SetFrequency(cpuID int, freqKHz uint64) error {
	// First, switch to userspace governor
	if err := m.SetGovernor(cpuID, GovernorUserspace); err != nil {
		return fmt.Errorf("failed to set userspace governor: %w", err)
	}

	cpuPath := filepath.Join(sysfsBasePath, fmt.Sprintf("cpu%d/cpufreq", cpuID))
	setspeedPath := filepath.Join(cpuPath, "scaling_setspeed")

	return writeString(setspeedPath, strconv.FormatUint(freqKHz, 10))
}
