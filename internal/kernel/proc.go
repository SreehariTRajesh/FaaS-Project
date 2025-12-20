package kernel

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
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
}

type ProcEvent = procProcEventT

func NewProcRuntimeMonitor(outputFilePath string) (*ProcRuntimeMonitor, error) {

	objs := &procObjects{}
	if err := loadProcObjects(objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load BPF objects: %w", err)
	}

	outputFile, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}

	outputFileWriter := csv.NewWriter(outputFile)

	outputFileWriter.Write([]string{
		"cgroup_id",
		"pid",
		"start_timestamp",
		"end_timestamp",
		"latency",
	})

	outputFileWriter.Flush()
	return &ProcRuntimeMonitor{
		links:            make([]link.Link, 0),
		outputFileWriter: outputFileWriter,
		objs:             objs,
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

func (p *ProcRuntimeMonitor) ReadEvents() error {
	for {
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
		})

		p.outputFileWriter.Flush()
	}
}

func (p *ProcRuntimeMonitor) UpdateContainerCgroupId(cgroupId uint32) error {
	err := p.objs.ProcessContainerMap.Update(cgroupId, true, ebpf.UpdateNoExist)
	if err != nil {
		return fmt.Errorf("error whiel updating cgroup id: %v", err)
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
