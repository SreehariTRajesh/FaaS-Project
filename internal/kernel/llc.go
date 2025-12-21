package kernel

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"golang.org/x/sys/unix"
)

type LLCMetricsMonitor struct {
	// Implementation details would go here
	outputFileWriter *csv.Writer
	objs             *llcObjects
	links            []link.Link
}

const (
	PERF_COUNT_HW_CACHE_LL            = 0x3
	PERF_COUNT_HW_CACHE_OP_READ       = 0x0 << 8
	PERF_COUNT_HW_CACHE_OP_WRITE      = 0x1 << 8
	PERF_COUNT_HW_CACHE_OP_PREFETCH   = 0x2 << 8
	PERF_COUNT_HW_CACHE_RESULT_ACCESS = 0x0 << 16
	PERF_COUNT_HW_CACHE_RESULT_MISS   = 0x1 << 16
)

func NewLLCMetricsMonitor(outputFilePath string) (*LLCMetricsMonitor, error) {

	// Implementation for loading BPF objects and setting up the monitor
	// would go here, similar to NewLatencyMetricsMonitor
	objs := &llcObjects{}

	if err := loadLlcObjects(objs, nil); err != nil {
		return nil, fmt.Errorf("loading eBPF objects: %w", err)
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
			"cgroup_id",
			"pid",
			"cpu",
			"read_hits",
			"read_misses",
			"read_references",
			"write_hits",
			"write_misses",
			"write_references",
			"prefetch_hits",
			"prefetch_misses",
			"prefetch_references",
			"total_hits",
			"total_misses",
			"total_references",
		})

		outputFileWriter.Flush()
	}
	return &LLCMetricsMonitor{
		links:            make([]link.Link, 0),
		outputFileWriter: outputFileWriter, // Placeholder
		objs:             objs,             // Placeholder
	}, nil
}

func (m *LLCMetricsMonitor) AttachLLCEvents() error {
	// Implementation for attaching BPF programs to relevant hooks
	numCPUs := runtime.NumCPU()
	events := []struct {
		config  uint64
		program *ebpf.Program
		name    string
	}{
		{PERF_COUNT_HW_CACHE_LL | PERF_COUNT_HW_CACHE_OP_READ | PERF_COUNT_HW_CACHE_RESULT_MISS, m.objs.LlcReadMissHandler, "llc_read_miss"},
		{PERF_COUNT_HW_CACHE_LL | PERF_COUNT_HW_CACHE_OP_READ | PERF_COUNT_HW_CACHE_RESULT_ACCESS, m.objs.LlcReadHitHandler, "llc_read_hit"},
		{PERF_COUNT_HW_CACHE_LL | PERF_COUNT_HW_CACHE_OP_WRITE | PERF_COUNT_HW_CACHE_RESULT_MISS, m.objs.LlcWriteMissHandler, "llc_write_miss"},
		{PERF_COUNT_HW_CACHE_LL | PERF_COUNT_HW_CACHE_OP_WRITE | PERF_COUNT_HW_CACHE_RESULT_ACCESS, m.objs.LlcWriteHitHandler, "llc_write_hit"},
		{PERF_COUNT_HW_CACHE_LL | PERF_COUNT_HW_CACHE_OP_PREFETCH | PERF_COUNT_HW_CACHE_RESULT_MISS, m.objs.LlcPrefetchMissHandler, "llc_prefetch_miss"},
		{PERF_COUNT_HW_CACHE_LL | PERF_COUNT_HW_CACHE_OP_PREFETCH | PERF_COUNT_HW_CACHE_RESULT_ACCESS, m.objs.LlcPrefetchHitHandler, "llc_prefetch_hit"},
	}

	for cpu := 0; cpu < numCPUs; cpu++ {
		for _, event := range events {
			fd, err := unix.PerfEventOpen(&unix.PerfEventAttr{
				Type:   unix.PERF_TYPE_HW_CACHE,
				Config: event.config,
				Size:   uint32(unsafe.Sizeof(unix.PerfEventAttr{})),
				Sample: 100,
				Bits:   unix.PerfBitDisabled | unix.PerfBitFreq,
			}, -1, cpu, -1, unix.PERF_FLAG_FD_CLOEXEC)

			if err != nil {
				continue
			}

			if err := unix.IoctlSetInt(fd, unix.PERF_EVENT_IOC_ENABLE, 0); err != nil {
				unix.Close(fd)
				log.Printf("Warning: failed to enable perf event %s on CPU %d: %v", event.name, cpu, err)
				continue
			}

			l, err := link.AttachRawLink(link.RawLinkOptions{
				Target:  fd,
				Program: event.program,
				Attach:  ebpf.AttachPerfEvent,
			})

			if err != nil {
				unix.Close(fd)
				return fmt.Errorf("attaching %s to CPU %d: %w", event.name, cpu, err)
			}

			m.links = append(m.links, l)
		}
	}

	return nil
}

func (m *LLCMetricsMonitor) GetStats(duration int32) error {
	// Implementation for reading events from the BPF maps/ring buffers
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ticker := time.NewTicker(1 * time.Millisecond)

	for time := 0; time < int(duration); time++ {

		select {
		case <-ticker.C:
			var key uint64
			var value llcLlcEventT
			iter := m.objs.LlcStatsMap.Iterate()
			for iter.Next(&key, &value) {
				m.outputFileWriter.Write([]string{
					strconv.FormatUint(uint64(value.CgroupId), 10),
					strconv.FormatUint(uint64(value.Pid), 10),
					strconv.FormatUint(uint64(value.Cpu), 10),
					strconv.FormatUint(uint64(value.ReadHits), 10),
					strconv.FormatUint(uint64(value.ReadMisses), 10),
					strconv.FormatUint(uint64(value.ReadReferences), 10),
					strconv.FormatUint(uint64(value.WriteHits), 10),
					strconv.FormatUint(uint64(value.WriteMisses), 10),
					strconv.FormatUint(uint64(value.WriteReferences), 10),
					strconv.FormatUint(uint64(value.PrefetchHits), 10),
					strconv.FormatUint(uint64(value.PrefetchMisses), 10),
					strconv.FormatUint(uint64(value.PrefetchReferences), 10),
					strconv.FormatUint(uint64(value.TotalHits), 10),
					strconv.FormatUint(uint64(value.TotalMisses), 10),
					strconv.FormatUint(uint64(value.TotalReferences), 10),
				})
			}

			if err := iter.Err(); err != nil {
				return fmt.Errorf("iterating map: %w", err)
			}
		case <-sig:
			fmt.Println("Shutting Down")
			return nil
		}

	}

	return nil
}

func (m *LLCMetricsMonitor) UpdateContainerCgroupId(cgroupId uint32) error {
	err := m.objs.LlcContainerMap.Update(cgroupId, true, ebpf.UpdateNoExist)
	if err != nil {
		return fmt.Errorf("error while updating cgroup id: %v", err)
	}

	return nil
}

func (m *LLCMetricsMonitor) Close() error {
	for _, l := range m.links {
		if err := l.Close(); err != nil {
			log.Printf("Warning: failed to close link: %v\n", err)
		}
	}

	if m.objs != nil {
		return m.objs.Close()
	}

	return nil
}
