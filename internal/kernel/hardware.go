package kernel

import (
	"encoding/csv"
	"errors"
	"faas-migration/internal/energy"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"golang.org/x/sys/unix"
)

type HardwareMetricsMonitor struct {
	outputFileWriter *csv.Writer
	objs             *hardwareObjects
	links            []link.Link
}

func NewHardwareMetricsMonitor(outputFilePath string) (*HardwareMetricsMonitor, error) {
	loadOpts := &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf/proc", // Pin directory
		},
	}

	objs := &hardwareObjects{}

	if err := loadHardwareObjects(objs, loadOpts); err != nil {
		return nil, fmt.Errorf("error while loading hardware objects: %w", err)
	}

	outputFile, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}

	outputFileWriter := csv.NewWriter(outputFile)

	outputFileInfo, err := os.Stat(outputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat output file: %w", err)
	}

	if outputFileInfo.Size() == 0 {
		outputFileWriter.Write([]string{
			"cycles",
			"instructions",
			"ref_cycles",
			"cache_references",
			"cache_misses",
			"branches",
			"branch_misses",
			"l1d_loads",
			"l1d_stores",
			"llc_loads",
			"llc_load_misses",
			"llc_stores",
			"llc_store_misses",
			"dtlb_loads",
			"dtlb_load_misses",
			"dtlb_stores",
			"dtlb_store_misses",
			"bpu_loads",
			"bpu_load_misses",
			"energy_uj",
		})
		outputFileWriter.Flush()
	}
	return &HardwareMetricsMonitor{
		links:            make([]link.Link, 0),
		outputFileWriter: outputFileWriter, // Placeholder
		objs:             objs,             // Placeholder
	}, nil
}

func (h *HardwareMetricsMonitor) UpdateContainerCgroupId(cgroupId uint32) error {
	err := h.objs.ProcStatsMap.Update(cgroupId, &hardwarePerfStats{}, ebpf.UpdateAny)
	if err != nil {
		return fmt.Errorf("error while updating cgroup id: %v", err)
	}
	return nil
}

func (h *HardwareMetricsMonitor) AttachHardwareEvents() error {
	numCPUs, err := ebpf.PossibleCPU()
	if err != nil {
		return fmt.Errorf("getting CPU count: %w", err)
	}

	// Define perf events to attach
	events := []struct {
		name       string
		prog       *ebpf.Program
		event_type int
		config     uint64
	}{
		// Hardware events (PERF_TYPE_HARDWARE)
		{
			name:       "CPU_CYCLES",
			prog:       h.objs.hardwarePrograms.OnCpuCycles,
			event_type: unix.PERF_TYPE_HARDWARE,
			config:     unix.PERF_COUNT_HW_CPU_CYCLES,
		},
		{
			name:       "INSTRUCTIONS",
			prog:       h.objs.hardwarePrograms.OnInstructions,
			event_type: unix.PERF_TYPE_HARDWARE,
			config:     unix.PERF_COUNT_HW_INSTRUCTIONS,
		},
		{
			name:       "REF_CYCLES",
			prog:       h.objs.hardwarePrograms.OnRefCycles,
			event_type: unix.PERF_TYPE_HARDWARE,
			config:     unix.PERF_COUNT_HW_REF_CPU_CYCLES,
		},
		{
			name:       "CACHE_MISSES",
			prog:       h.objs.hardwarePrograms.OnCacheMisses,
			event_type: unix.PERF_TYPE_HARDWARE,
			config:     unix.PERF_COUNT_HW_CACHE_MISSES,
		},
		{
			name:       "CACHE_REFERENCES",
			prog:       h.objs.hardwarePrograms.OnCacheReferences,
			event_type: unix.PERF_TYPE_HARDWARE,
			config:     unix.PERF_COUNT_HW_CACHE_REFERENCES,
		},
		{
			name:       "BRANCHES",
			prog:       h.objs.hardwarePrograms.OnBranches,
			event_type: unix.PERF_TYPE_HARDWARE,
			config:     unix.PERF_COUNT_HW_BRANCH_INSTRUCTIONS,
		},
		{
			name:       "BRANCH_MISSES",
			prog:       h.objs.hardwarePrograms.OnBranchMisses,
			event_type: unix.PERF_TYPE_HARDWARE,
			config:     unix.PERF_COUNT_HW_BRANCH_MISSES,
		},

		// Hardware cache events (PERF_TYPE_HW_CACHE)
		{
			name:       "L1D_LOADS",
			prog:       h.objs.hardwarePrograms.OnL1dLoads,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_L1D | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16),
		},
		{
			name:       "L1D_STORES",
			prog:       h.objs.hardwarePrograms.OnL1dStores,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_L1D | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16),
		},
		{
			name:       "LLC_LOADS",
			prog:       h.objs.hardwarePrograms.OnLlcLoads,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_LL | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16),
		},
		{
			name:       "LLC_LOAD_MISSES",
			prog:       h.objs.hardwarePrograms.OnLlcLoadMisses,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_LL | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16),
		},
		{
			name:       "LLC_STORES",
			prog:       h.objs.hardwarePrograms.OnLlcStores,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_LL | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16),
		},
		{
			name:       "LLC_STORE_MISSES",
			prog:       h.objs.hardwarePrograms.OnLlcStoreMisses,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_LL | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16),
		},
		{
			name:       "DTLB_LOADS",
			prog:       h.objs.hardwarePrograms.OnDtlbLoads,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_DTLB | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16),
		},
		{
			name:       "DTLB_LOAD_MISSES",
			prog:       h.objs.hardwarePrograms.OnDtlbLoads,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_DTLB | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16),
		},
		{
			name:       "DTLB_STORES",
			prog:       h.objs.hardwarePrograms.OnDtlbStores,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_DTLB | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16),
		},
		{
			name:       "DTLB_STORE_MISSES",
			prog:       h.objs.hardwarePrograms.OnDtlbStoreMisses,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_DTLB | (unix.PERF_COUNT_HW_CACHE_OP_WRITE << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16),
		},
		{
			name:       "BPU_LOADS",
			prog:       h.objs.hardwarePrograms.OnBpuLoads,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_BPU | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS << 16),
		},
		{
			name:       "BPU_LOAD_MISSES",
			prog:       h.objs.hardwarePrograms.OnBpuLoadMisses,
			event_type: unix.PERF_TYPE_HW_CACHE,
			config:     unix.PERF_COUNT_HW_CACHE_BPU | (unix.PERF_COUNT_HW_CACHE_OP_READ << 8) | (unix.PERF_COUNT_HW_CACHE_RESULT_MISS << 16),
		},
	}

	// Attach to each CPU
	for _, event := range events {
		for cpu := 0; cpu < numCPUs; cpu++ {
			attr := unix.PerfEventAttr{
				Type:   uint32(event.event_type),
				Config: event.config,
				Bits:   unix.PerfBitDisabled,
				Sample: 1000,
				Size:   uint32(unsafe.Sizeof(unix.PerfEventAttr{})),
			}

			// Open perf event
			fd, err := unix.PerfEventOpen(
				&attr,
				-1,  // PID (-1 for all processes)
				cpu, // CPU
				-1,  // Group FD
				unix.PERF_FLAG_FD_CLOEXEC,
			)

			if err != nil {
				if errors.Is(err, unix.ENOENT) || errors.Is(err, unix.EOPNOTSUPP) {
					fmt.Printf("Warning: %s not supported on CPU %d, skipping\n", event.name, cpu)
					continue
				}
				return fmt.Errorf("opening %s perf event on CPU %d: %w", event.name, cpu, err)
			}

			// ENABLE the perf event - THIS IS CRITICAL
			err = unix.IoctlSetInt(fd, unix.PERF_EVENT_IOC_ENABLE, 0)
			if err != nil {
				unix.Close(fd)
				return fmt.Errorf("enabling perf event: %w", err)
			}

			// Attach eBPF program to perf event
			l, err := link.AttachRawLink(link.RawLinkOptions{
				Target:  fd,
				Program: event.prog,
				Attach:  ebpf.AttachPerfEvent,
			})

			if err != nil {
				unix.Close(fd)
				return fmt.Errorf("attaching %s program to CPU %d: %w", event.name, cpu, err)
			}

			h.links = append(h.links, l)
		}
		fmt.Printf("Attached %s to %d CPUs\n", event.name, numCPUs)
	}

	return nil
}

func (h *HardwareMetricsMonitor) ReadStats(interval time.Duration) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// FIX: If 'interval' is already time.Duration, do not multiply by Millisecond.
	// If 'interval' is an int representing ms, cast it: time.Duration(interval) * time.Millisecond
	ticker := time.NewTicker(interval * time.Millisecond)
	defer ticker.Stop()

	// Initialize previous values
	var (
		currEnergy uint64
		prevEnergy uint64
		currStats  hardwarePerfStats
		energyErr  error
		statsErr   error
		wg         sync.WaitGroup
		statsKey   uint32 = 0
		prevStats  hardwarePerfStats
		prevKey    uint32 = 0
	)

	wg.Add(2)

	// 1. Read Energy
	go func() {
		defer wg.Done()
		prevEnergy, energyErr = energy.ReadRAPLEnergyUJ()
	}()

	// 2. Read Stats
	go func() {
		defer wg.Done()
		statsErr = h.objs.hardwareMaps.ProcStatsMap.Lookup(&statsKey, &prevStats)
	}()

	// Wait for both to finish
	wg.Wait()

	// Initial Reads
	if err := h.objs.hardwareMaps.ProcStatsMap.Lookup(&prevKey, &prevStats); err != nil {
		return fmt.Errorf("error reading initial stats: %w", err)
	}
	prevEnergy, err := energy.ReadRAPLEnergyUJ()
	if err != nil {
		return fmt.Errorf("error while reading initial energy: %w", err)
	}

	// Reusable variables to reduce allocation overhead in the loop

	for {
		select {
		case <-ticker.C:
			// --- CRITICAL SECTION START ---
			// We spawn two goroutines to minimize the time delta (skew)
			// between reading Energy and reading Stats.

			wg.Add(2)

			// 1. Read Energy
			go func() {
				defer wg.Done()
				currEnergy, energyErr = energy.ReadRAPLEnergyUJ()
			}()

			// 2. Read Stats
			go func() {
				defer wg.Done()
				statsErr = h.objs.hardwareMaps.ProcStatsMap.Lookup(&statsKey, &currStats)
			}()

			// Wait for both to finish
			wg.Wait()
			// --- CRITICAL SECTION END ---

			// Handle errors after the critical timing window
			if energyErr != nil {
				return fmt.Errorf("error while reading energy: %w", energyErr)
			}
			if statsErr != nil {
				return fmt.Errorf("error reading stats: %w", statsErr)
			}

			// Calculate Diffs
			energyDiff := currEnergy - prevEnergy

			// Write output
			// Optimization: Using strconv.AppendUint is slightly faster/more efficient
			// than FormatUint for high-frequency loops, but FormatUint is fine here.
			h.outputFileWriter.Write([]string{
				strconv.FormatUint(currStats.Cycles-prevStats.Cycles, 10),
				strconv.FormatUint(currStats.Instructions-prevStats.Instructions, 10),
				strconv.FormatUint(currStats.RefCycles-prevStats.RefCycles, 10),
				strconv.FormatUint(currStats.CacheReferences-prevStats.CacheReferences, 10),
				strconv.FormatUint(currStats.CacheMisses-prevStats.CacheMisses, 10),
				strconv.FormatUint(currStats.Branches-prevStats.Branches, 10),
				strconv.FormatUint(currStats.BranchMisses-prevStats.BranchMisses, 10),
				strconv.FormatUint(currStats.L1dLoads-prevStats.L1dLoads, 10),
				strconv.FormatUint(currStats.L1dStores-prevStats.L1dStores, 10),
				strconv.FormatUint(currStats.LlcLoads-prevStats.LlcLoads, 10),
				strconv.FormatUint(currStats.LlcLoadMisses-prevStats.LlcLoadMisses, 10),
				strconv.FormatUint(currStats.LlcStores-prevStats.LlcStores, 10),
				strconv.FormatUint(currStats.LlcStoreMisses-prevStats.LlcStoreMisses, 10),
				strconv.FormatUint(currStats.DtlbLoads-prevStats.DtlbLoads, 10),
				strconv.FormatUint(currStats.DtlbLoadMisses-prevStats.DtlbLoadMisses, 10),
				strconv.FormatUint(currStats.DtlbStores-prevStats.DtlbStores, 10),
				strconv.FormatUint(currStats.DtlbStoreMisses-prevStats.DtlbStoreMisses, 10),
				strconv.FormatUint(currStats.BpuLoads-prevStats.BpuLoads, 10),
				strconv.FormatUint(currStats.BpuLoadMisses-prevStats.BpuLoadMisses, 10),
				strconv.FormatUint(energyDiff, 10),
			})
			h.outputFileWriter.Flush()

			// Update previous values for next iteration
			prevEnergy = currEnergy
			prevStats = currStats

		case <-sig:
			fmt.Println("Shutting Down")
			return nil
		}
	}
}

func (h *HardwareMetricsMonitor) InitializeStats() error {
	err := h.objs.hardwareMaps.ProcStatsMap.Update(0, hardwarePerfStats{
		Cycles:          0,
		Instructions:    0,
		CacheMisses:     0,
		CacheReferences: 0,
		BranchMisses:    0,
	}, ebpf.UpdateNoExist)

	if err != nil {
		return fmt.Errorf("error while initializing stats: %w", err)
	}

	return nil
}

func (h *HardwareMetricsMonitor) Close() error {
	for _, link := range h.links {
		if err := link.Close(); err != nil {
			return err
		}
	}

	return h.objs.Close()
}
