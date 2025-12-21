package kernel

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

type FunctionMetricsMonitor struct {
	outputFileWriter *csv.Writer
	objs             *functionObjects
	links            []link.Link
	reader           *ringbuf.Reader
}

type FunctionEvent struct {
	Pid        int32
	_          [4]byte
	DurationNs int64
}

func NewFunctionMetricsMonitor(outputFilePath string) (*FunctionMetricsMonitor, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock: %w", err)
	}

	objs := &functionObjects{}
	if err := loadFunctionObjects(objs, nil); err != nil {
		return nil, fmt.Errorf("failed to load BPF objects: %w", err)
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
			"pid",
			"duration",
		})

		outputFileWriter.Flush()
	}
	return &FunctionMetricsMonitor{
		outputFileWriter: outputFileWriter,
		objs:             objs,
		links:            make([]link.Link, 0),
		reader:           nil,
	}, nil
}

func (f *FunctionMetricsMonitor) Attach(binaryPath *string, symbol *string) error {
	ex, err := link.OpenExecutable(*binaryPath)
	if err != nil {
		return fmt.Errorf("error opening executable: %v", err)
	}

	upEntry, err := ex.Uprobe(*symbol, f.objs.UprobeEntry, nil)
	if err != nil {
		return fmt.Errorf("error attaching uprobe: %v", err)
	}

	f.links = append(f.links, upEntry)

	upExit, err := ex.Uretprobe(*symbol, f.objs.UprobeExit, nil)
	if err != nil {
		return fmt.Errorf("error attaching uprobe: %v", err)
	}

	f.links = append(f.links, upExit)

	reader, err := ringbuf.NewReader(f.objs.FunctionEvents)

	if err != nil {
		return fmt.Errorf("error while opening ring buffer reader: %v", err)
	}

	f.reader = reader

	return nil
}

func (f *FunctionMetricsMonitor) ReadEvents(done chan struct{}) error {
	record, err := f.reader.Read()

	if err != nil {
		if errors.Is(err, ringbuf.ErrClosed) {
			return nil
		}
		return fmt.Errorf("reading from ringbuf: %w", err)
	}

	var event FunctionEvent
	if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &event); err != nil {
		return fmt.Errorf("failed to parse event: %v", err)
	}

	fmt.Println(event)

	f.outputFileWriter.Write([]string{
		strconv.FormatUint(uint64(event.Pid), 10),
		strconv.FormatUint(uint64(event.DurationNs), 10),
	})

	f.outputFileWriter.Flush()

	close(done)
	return nil
}

func (f *FunctionMetricsMonitor) Close() error {
	if f.reader != nil {
		if err := f.reader.Close(); err != nil {
			return err
		}
	}

	for _, link := range f.links {
		if err := link.Close(); err != nil {
			return err
		}
	}

	return f.objs.Close()
}
