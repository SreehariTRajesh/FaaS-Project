package kernel

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"faas-migration/internal/energy"
	"fmt"
	"os"
	"strconv"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

type ProcRuntimeMonitor struct {
	outputFileWriter *csv.Writer
	objs             *procObjects
	links            []link.Link
	reader           *ringbuf.Reader
	msrManager       *energy.MSRManager
}

type ProcEvent = procProcEventT

func NewProcRuntimeMonitor(outputFilePath string) (*ProcRuntimeMonitor, error) {
	loadOpts := &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf/proc", // Pin directory
		},
	}
	objs := &procObjects{}
	if err := loadProcObjects(objs, loadOpts); err != nil {
		return nil, fmt.Errorf("failed to load BPF objects: %w", err)
	}

	outputFile, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}

	msrManager := energy.NewMSRManager()

	outputFileWriter := csv.NewWriter(outputFile)

	outputFileInfo, err := os.Stat(outputFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat output file: %w", err)
	}

	if outputFileInfo.Size() == 0 {
		outputFileWriter.Write([]string{
			"cgroup_id",
			"pid",
			"start_timestamp",
			"end_timestamp",
			"latency",
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
		})
		outputFileWriter.Flush()
	}

	return &ProcRuntimeMonitor{
		links:            make([]link.Link, 0),
		outputFileWriter: outputFileWriter,
		objs:             objs,
		msrManager:       msrManager,
	}, nil
}

func (p *ProcRuntimeMonitor) Attach() error {
	tp, err := link.Tracepoint("syscalls", "sys_enter_execve", p.objs.TraceExecve, nil)

	if err != nil {
		return fmt.Errorf("error while attaching tracepoint: %v", err)
	}

	p.links = append(p.links, tp)

	kp, err := link.Kprobe("do_exit", p.objs.KprobeDoExit, nil)

	if err != nil {
		return fmt.Errorf("error whiel attaching tracepoint: %v", err)
	}

	p.links = append(p.links, kp)

	reader, err := ringbuf.NewReader(p.objs.Events)

	if err != nil {
		return fmt.Errorf("failed to create ringbuf reader: %w", err)
	}

	p.reader = reader

	return nil
}

func (p *ProcRuntimeMonitor) ReadEvents(done chan struct{}) error {
	record, err := p.reader.Read()
	if err != nil {
		if errors.Is(err, ringbuf.ErrClosed) {
			return nil
		}
		return fmt.Errorf("reading from ringbuf: %w", err)
	}
	var event ProcEvent

	if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
		return fmt.Errorf("failed to parse event: %v", err)
	}

	p.outputFileWriter.Write([]string{
		strconv.FormatUint(uint64(event.CgroupId), 10),
		strconv.FormatUint(uint64(event.Latency), 10),
		strconv.FormatUint(uint64(event.Pid), 10),
		strconv.FormatUint(uint64(event.StartTimestamp), 10),
		strconv.FormatUint(uint64(event.EndTimestamp), 10),
		strconv.FormatUint(event.HwStats.Cycles, 10),
		strconv.FormatUint(event.HwStats.Instructions, 10),
		strconv.FormatUint(event.HwStats.RefCycles, 10),
		strconv.FormatUint(event.HwStats.CacheReferences, 10),
		strconv.FormatUint(event.HwStats.CacheMisses, 10),
		strconv.FormatUint(event.HwStats.Branches, 10),
		strconv.FormatUint(event.HwStats.BranchMisses, 10),
		strconv.FormatUint(event.HwStats.L1dLoads, 10),
		strconv.FormatUint(event.HwStats.L1dStores, 10),
		strconv.FormatUint(event.HwStats.LlcLoads, 10),
		strconv.FormatUint(event.HwStats.LlcLoadMisses, 10),
		strconv.FormatUint(event.HwStats.LlcStores, 10),
		strconv.FormatUint(event.HwStats.LlcStoreMisses, 10),
		strconv.FormatUint(event.HwStats.DtlbLoads, 10),
		strconv.FormatUint(event.HwStats.DtlbLoadMisses, 10),
		strconv.FormatUint(event.HwStats.DtlbStores, 10),
		strconv.FormatUint(event.HwStats.DtlbStoreMisses, 10),
		strconv.FormatUint(event.HwStats.BpuLoads, 10),
		strconv.FormatUint(event.HwStats.BpuLoadMisses, 10),
	})

	p.outputFileWriter.Flush()
	close(done)
	return nil
}

func (p *ProcRuntimeMonitor) UpdateContainerCgroupId(cgroupId uint32) error {
	err := p.objs.ProcessContainerMap.Update(cgroupId, true, ebpf.UpdateAny)
	if err != nil {
		return fmt.Errorf("error while updating cgroup id: %v", err)
	}
	return nil
}

func (p *ProcRuntimeMonitor) Close() error {
	if p.reader != nil {
		if err := p.reader.Close(); err != nil {
			return err
		}
	}

	for _, link := range p.links {
		if err := link.Close(); err != nil {
			return err
		}
	}

	return p.objs.Close()
}
