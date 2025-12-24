package energy

import (
	"encoding/binary"
	"fmt"
	"os"
)

type MSRManager struct {
}

func NewMSRManager() *MSRManager {
	return &MSRManager{}
}

func (m *MSRManager) ReadCPUCoreEnergy(core int) (uint64, error) {
	path := fmt.Sprintf("/dev/cpu/%d/msr", core)

	// 2. Open the file in read-only mode
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open MSR file: %w", err)
	}
	defer f.Close()

	// 3. Read 8 bytes (64 bits) starting at the MSR address offset
	val := make([]byte, 8)
	_, err = f.ReadAt(val, int64(0x611))
	if err != nil {
		return 0, fmt.Errorf("failed to read MSR at 0x611: %w", err)
	}

	// 4. Convert byte slice to uint64 (MSRs use Little Endian)
	return binary.LittleEndian.Uint64(val), nil
}
