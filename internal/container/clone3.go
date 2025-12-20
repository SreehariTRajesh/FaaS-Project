package container

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const SYS_CLONE3 = 435

// CloneArgs is a representation of the clone3_args struct from Linux kernel.
type CloneArgs struct {
	Flags      uint64
	Pidfd      uint64
	ChildTid   uint64
	ParentTid  uint64
	ExitSignal uint64
	Stack      uint64
	StackSize  uint64
	TLS        uint64
	SetTid     uint64
	SetTidSize uint64
	Cgroup     uint64
	_          [16]byte // Padding to match the size of the struct in C
}

type Clone3Executor struct {
	// Add fields if necessary
	cgroupFd   int
	cgroupPath string
}

func NewClone3Executor(cgroupPath string) (*Clone3Executor, error) {

	cgroupFd, err := syscall.Open(cgroupPath, syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open cgroup path %s: %w", cgroupPath, err)
	}

	return &Clone3Executor{
		cgroupFd:   cgroupFd,
		cgroupPath: cgroupPath,
	}, nil
}

func (e *Clone3Executor) CloneIntoCgroup(binary string) uint32 {
	cloneArgs := &CloneArgs{
		Flags:      syscall.CLONE_INTO_CGROUP,
		ExitSignal: uint64(syscall.SIGCHLD),
		Cgroup:     uint64(e.cgroupFd),
	}

	fmt.Println("Prepared clone args calling clone3 syscall")
	pid, _, errno := syscall.Syscall(SYS_CLONE3, uintptr(unsafe.Pointer(cloneArgs)), unsafe.Sizeof(*cloneArgs), 0)
	if errno != 0 {
		fmt.Printf("clone3 syscall failed: %v\n", errno)
		os.Exit(1)
	}

	if pid == 0 {
		argv := []string{binary} // Use the path as the first argument

		// syscall.Exec REPLACES the current process with python3
		// This process ID remains the same, but the code becomes Python.
		err := syscall.Exec(binary, argv, os.Environ())

		// This code runs ONLY if Exec fails
		if err != nil {
			fmt.Printf("Failed to exec: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Python script executed successfully in child process")
		os.Exit(0)
	}

	return uint32(pid)
}

func (e *Clone3Executor) Close() error {
	if e.cgroupFd > 0 {
		return syscall.Close(e.cgroupFd)
	}
	return nil
}
