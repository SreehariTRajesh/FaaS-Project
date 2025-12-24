package energy

import (
	"os"
	"strconv"
	"strings"
)

const coreEnergyPath = "/sys/class/powercap/intel-rapl:0/energy_uj"

func ReadRAPLEnergyUJ() (uint64, error) {
	data, err := os.ReadFile(coreEnergyPath)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}
