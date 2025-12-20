import os
import sys
import ctypes
import ctypes.util
from ctypes import c_int, c_uint64, c_size_t, c_void_p, POINTER, Structure
import subprocess
import time
import argparse

CLONE_NEWCGROUP = 0x02000000
CLONE_INTO_CGROUP = 0x200000000
CSIGNAL = 0x000000FF


class clone_args(Structure):
    _fields_ = [
        ("flags", c_uint64),
        ("pidfd", c_uint64),
        ("child_tid", c_uint64),
        ("parent_tid", c_uint64),
        ("exit_signal", c_uint64),
        ("stack", c_uint64),
        ("stack_size", c_uint64),
        ("tls", c_uint64),
        ("set_tid", c_uint64),
        ("set_tid_size", c_size_t),
        ("cgroup", c_uint64),
    ]


class CgroupManager:
    def __init__(self, cgroup_name: str, cgroup_root: str = "/sys/fs/cgroup"):
        self.cgroup_name = cgroup_name
        self.cgroup_root = cgroup_root
        self.cgroup_path = os.path.join(cgroup_root, cgroup_name)

    def create_cgroup(self) -> str:
        """Create a new cgroup"""
        try:
            os.makedirs(self.cgroup_path, exist_ok=True)
            print(f"[Cgroup] Created cgroup at: {self.cgroup_path}")
            return self.cgroup_path
        except PermissionError:
            print("[ERROR] Need root privileges to create cgroups")
            sys.exit(1)

    def set_cpuset(self, cpus: str):
        """Set CPU affinity for the cgroup"""
        cpuset_file = os.path.join(self.cgroup_path, "cpuset.cpus")
        try:
            with open(cpuset_file, "w") as f:
                f.write(cpus)
            print(f"[Cgroup] Set cpuset.cpus to: {cpus}")
        except Exception as e:
            print(f"[ERROR] Failed to set cpuset: {e}")

    def set_memory_limit(self, limit: str):
        """Set memory limit (e.g., '512M', '1G')"""
        memory_file = os.path.join(self.cgroup_path, "memory.max")
        try:
            with open(memory_file, "w") as f:
                f.write(limit)
            print(f"[Cgroup] Set memory.max to: {limit}")
        except Exception as e:
            print(f"[ERROR] Failed to set memory limit: {e}")

    def get_cgroup_fd(self) -> int:
        """Get file descriptor for the cgroup directory"""
        return os.open(self.cgroup_path, os.O_DIRECTORY | os.O_RDONLY)


    def get_cgroup_id(self) -> int:
        try: 
            stat_info = os.stat(self.cgroup_path)
            return stat_info.st_ino
        except Exception as e:
            print(f"Error getting inode: {e}")
            return None
    
    def cleanup(self):
        """Remove the cgroup"""
        try:
            os.rmdir(self.cgroup_path)
            print(f"[Cgroup] Removed cgroup: {self.cgroup_path}")
        except Exception as e:
            print(f"[WARNING] Failed to remove cgroup: {e}")


class Clone3Runner:

    def __init__(self):
        libc_path = ctypes.util.find_library("c")
        self.libc = ctypes.CDLL(libc_path, use_errno=True)

        self.SYS_clone3 = 435
        self.libc.syscall.argtypes = [c_int, c_void_p, c_size_t]
        self.libc.syscall.restype = c_int

    def clone3_into_cgroup(self, cgroup_fd: int, command: list) -> int:
        args = clone_args()
        args.flags = CLONE_INTO_CGROUP
        args.exit_signal = 17  # SIGCHLD
        args.cgroup = cgroup_fd

        # Fork using clone3
        pid = self.libc.syscall(
            self.SYS_clone3, ctypes.byref(args), ctypes.sizeof(args)
        )

        if pid == -1:
            errno = ctypes.get_errno()
            raise OSError(errno, f"clone3 failed: {os.strerror(errno)}")

        if pid == 0:
            # Child process - exec the command
            os.execvp(command[0], command)
            sys.exit(1)

        return pid


def run_benchmark_in_cgroup(
    benchmark_script: str,
    cgroup_name: str,
    cpuset: str,
    memory_limit: str,
    duration: float = 60.0,
):
    cgroup_manager = CgroupManager(cgroup_name=cgroup_name)
    cgroup_path = cgroup_manager.create_cgroup()

    try:
        cgroup_manager.set_cpuset(cpuset)
        cgroup_manager.set_memory_limit(memory_limit)

        cgroup_fd = cgroup_manager.get_cgroup_fd()

        print("Cgroup ID (inode):", cgroup_manager.get_cgroup_id())

        command = [sys.executable, benchmark_script]
        print(f"[Runner] Running command: {' '.join(command)}")
        print("Waiting for input to start benchmark...")
        input()

        runner = Clone3Runner()
        child_pid = runner.clone3_into_cgroup(cgroup_fd=cgroup_fd, command=command)

        print(f"[Runner] Child process started with PID: {child_pid}")

        # Close the cgroup fd in parent
        os.close(cgroup_fd)
        print(f"Start Migration Now to CPU?")
        cpuset = input()

        cgroup_manager.set_cpuset(cpuset)

        print("Benchmark running... Press Ctrl+C to stop early.")
        # Wait for child process
        start_time = time.time()
        pid, status = os.waitpid(child_pid, 0)
        elapsed = time.time() - start_time

        exit_code = os.WEXITSTATUS(status)
        print(f"\n[Runner] Process {pid} exited with code {exit_code}")
        print(f"[Runner] Total execution time: {elapsed:.2f}s")

        # Read cgroup stats

    except Exception as e:
        print(f"[ERROR] {e}")
        import traceback

        traceback.print_exc()
    finally:
        # Cleanup
        time.sleep(0.5)  # Give time for cgroup to settle
        cgroup_manager.cleanup()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run Serverless Benchmark in custom cgroup"
    )

    parser.add_argument("--benchmark-script", help="Path to benchmark python script")

    parser.add_argument(
        "--cgroup-name", default="serverless_benchmark", help="Name of cgroup to create"
    )

    parser.add_argument(
        "--cpuset", default="0", help='CPU set (e.g., "0", "0-3", "0,2,4")'
    )
    parser.add_argument(
        "--memory", default="1G", help='Memory limit (e.g., "512M", "1G")'
    )
    parser.add_argument(
        "--duration", type=float, default=60.0, help="Benchmark duration in seconds"
    )

    args = parser.parse_args()

    run_benchmark_in_cgroup(
        benchmark_script=args.benchmark_script,
        cgroup_name=args.cgroup_name,
        cpuset=args.cpuset,
        memory_limit=args.memory,
        duration=args.duration
    )
