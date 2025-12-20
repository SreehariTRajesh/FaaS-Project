package kernel

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

type LatencyMetricsMonitor struct {
	outputFileWriter *csv.Writer
	objs             *latencyObjects
	links            []link.Link
	reader           *ringbuf.Reader
}

type MigrationEvent = latencyMigrationEventT

func NewLatencyMetricsMonitor(outputFilePath string) (*LatencyMetricsMonitor, error) {

	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock: %w", err)
	}

	objs := &latencyObjects{}
	if err := loadLatencyObjects(objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load BPF objects: %w", err)
	}

	outputFile, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}

	outputFileWriter := csv.NewWriter(outputFile)

	outputFileWriter.Write([]string{
		"cgroup_id",
		"latency",
		"pid",
		"source_cpu",
		"target_cpu",
	})

	outputFileWriter.Flush()

	return &LatencyMetricsMonitor{
		objs:             objs,
		links:            make([]link.Link, 0),
		outputFileWriter: outputFileWriter,
	}, nil
}

func (m *LatencyMetricsMonitor) Attach() error {
	rtpMigrate, err := link.AttachRawTracepoint(link.RawTracepointOptions{
		Name:    "sched_migrate_task",
		Program: m.objs.TracepointSchedMigrateTask,
	})

	if err != nil {
		return fmt.Errorf("failed to attach sched_migrate_task: %w", err)
	}

	log.Println("raw tracepoint sched_migrate_task attached successfully")

	m.links = append(m.links, rtpMigrate)

	rtpSwitch, err := link.AttachRawTracepoint(link.RawTracepointOptions{
		Name:    "sched_switch",
		Program: m.objs.RawTpSchedSwitch,
	})

	if err != nil {
		return fmt.Errorf("failed to attach sched_switch: %w", err)
	}

	m.links = append(m.links, rtpSwitch)

	log.Println("raw tracepoint sched_switch attached successfully")

	reader, err := ringbuf.NewReader(m.objs.Events)

	if err != nil {
		return fmt.Errorf("failed to create ringbuf reader: %w", err)
	}
	m.reader = reader

	log.Println("migration monitor attached successfully")
	return nil
}

func (m *LatencyMetricsMonitor) ReadEvents(done chan struct{}) error {
	record, err := m.reader.Read()
	if err != nil {
		if errors.Is(err, ringbuf.ErrClosed) {
			return nil
		}
		return fmt.Errorf("reading from ringbuf: %w", err)
	}
	var event MigrationEvent

	if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
		return fmt.Errorf("failed to parse event: %v", err)
	}

	m.outputFileWriter.Write([]string{
		strconv.FormatUint(uint64(event.CgroupId), 10),
		strconv.FormatUint(event.Latency, 10),
		strconv.FormatUint(uint64(event.Pid), 10),
		strconv.FormatUint(uint64(event.SourceCpu), 10),
		strconv.FormatUint(uint64(event.TargetCpu), 10),
	})
	m.outputFileWriter.Flush()

	close(done)
	return nil
}

// GetEventChannel returns the event channel for reading events

func (m *LatencyMetricsMonitor) UpdateContainerCgroupId(cgroupId uint32) error {
	err := m.objs.ContainerMap.Update(cgroupId, true, ebpf.UpdateNoExist)
	if err != nil {
		return fmt.Errorf("error while updating cgroup id: %v", err)
	}

	fmt.Println("successfully updated container cgroup id:", cgroupId)

	return nil
}

func (m *LatencyMetricsMonitor) Close() error {
	if m.reader != nil {
		if err := m.reader.Close(); err != nil {
			return err
		}
	}

	for _, link := range m.links {
		if err := link.Close(); err != nil {
			return err
		}
	}

	return m.objs.Close()
}
