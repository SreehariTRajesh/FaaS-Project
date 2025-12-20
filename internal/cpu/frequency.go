package cpu

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	sysfsBasePath = "/sys/devices/system/cpu"
)

type CPUFrequencyManager struct {
	numCPUS int
}

type CPUInfo struct {
	CPUId              int
	CurrFreq           uint64
	MinFreq            uint64
	MaxFreq            uint64
	ScalingMinFreq     uint64
	ScalingMaxFreq     uint64
	Governor           string
	AvailableGovernors []string
	Driver             string
}

func NewCPUFreqManager() (*CPUFrequencyManager, error) {
	numCPUs, err := getNumCPUs()
	if err != nil {
		return nil, fmt.Errorf("failed to get number of CPUs: %w", err)
	}

	return &CPUFrequencyManager{
		numCPUS: numCPUs,
	}, nil
}

func getNumCPUs() (int, error) {
	entries, err := os.ReadDir(sysfsBasePath)

	if err != nil {
		return 0, err
	}

	count := 0

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cpu") && entry.IsDir() {
			cpuNumStr := strings.TrimPrefix(entry.Name(), "cpu")
			if _, err := strconv.Atoi(cpuNumStr); err == nil {
				count++
			}
		}
	}

	return count, nil
}

func (m *CPUFrequencyManager) GetNumCPUs() int {
	return m.numCPUS
}

func (m *CPUFrequencyManager) GetCPUInfo(cpuId int) (*CPUInfo, error) {
	if cpuId < 0 || cpuId >= m.numCPUS {
		return nil, fmt.Errorf("invalid CPU Id: %d", cpuId)
	}

	cpuPath := filepath.Join(sysfsBasePath, fmt.Sprintf("cpu%d/cpufreq", cpuId))

	if _, err := os.Stat(cpuPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cpufreq not available for CPU: %d", cpuId)
	}

	currFreq, err := readUint64(filepath.Join(cpuPath, "scaling_cur_freq"))

	if err != nil {
		return nil, fmt.Errorf("failed to read current frequency: %w", err)
	}

	minFreq, err := readUint64(filepath.Join(cpuPath, "cpuinfo_min_freq"))

	if err != nil {
		return nil, fmt.Errorf("failed to read current frequency: %w", err)
	}

	maxFreq, err := readUint64(filepath.Join(cpuPath, "cpuinfo_max_freq"))

	if err != nil {
		return nil, fmt.Errorf("failed to read current frequency: %w", err)
	}

	scalingMinFreq, err := readUint64(filepath.Join(cpuPath, "scaling_min_freq"))

	if err != nil {
		return nil, fmt.Errorf("failed to read current frequency: %w", err)
	}

	scalingMaxFreq, err := readUint64(filepath.Join(cpuPath, "scaling_max_freq"))

	if err != nil {
		return nil, fmt.Errorf("failed to read current frequency: %w", err)
	}

	governor, err := readString(filepath.Join(cpuPath, "scaling_governor"))
	if err != nil {
		return nil, fmt.Errorf("failed to read governor: %w", err)
	}

	// Read available governors
	availGov, err := readString(filepath.Join(cpuPath, "scaling_available_governors"))
	if err != nil {
		return nil, fmt.Errorf("failed to read available governor: %w", err)
	}

	// Read driver
	driver, err := readString(filepath.Join(cpuPath, "scaling_driver"))
	if err != nil {
		return nil, fmt.Errorf("failed to read driver: %w", err)
	}

	return &CPUInfo{
		CPUId:              cpuId,
		CurrFreq:           currFreq,
		MinFreq:            minFreq,
		MaxFreq:            maxFreq,
		ScalingMinFreq:     scalingMinFreq,
		ScalingMaxFreq:     scalingMaxFreq,
		Governor:           strings.TrimSpace(governor),
		AvailableGovernors: strings.Fields(availGov),
		Driver:             strings.TrimSpace(driver),
	}, nil
}

func readUint64(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func readString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeString(path, value string) error {
	return os.WriteFile(path, []byte(value), 0644)
}
