import subprocess
import os
import itertools
import time

# --- CONFIGURATION VARIABLES ---

# Path to your compiled Go benchmark runner executable
GO_BENCHMARK_RUNNER = "./bin/cli"

# List of benchmark files (Python scripts) to test
BENCHMARK_FILES = [
    "benchmarks/cnnserv.py",
    "benchmarks/imagepr.py",
    "benchmarks/linpack.py",
    "benchmarks/lrserv.py",
    "benchmarks/rnnserv.py",
    "benchmarks/vidpr.py",
    "benchmarks/webserv.py",
    "benchmarks/wordcnt.py"
]

# List of frequency pairs (Old CPU Freq, New CPU Freq) in Hz
# The values are examples based on your previous input.
FREQUENCIES = [1000000, 1500000, 2000000, 2500000, 3000000, 3500000, 4000000, 4500000, 5000000]
FREQUENCY_PAIRS = [
    (old, new) for old in FREQUENCIES for new in FREQUENCIES
]

existing_pairs = [
]

# List of core migration scenarios (Old CPUSet, New CPUSet)
CORE_MIGRATIONS = [
    ("0", "1"),
]

# Number of times to repeat the entire configuration loop
RUNS_PER_CONFIG = 100

# Static Cgroup Parameters (usually don't change per run)
CGROUP_MEMORY_LIMIT = "256M"
BASE_CGROUP_NAME = "faas-benchmark"
LATENCY_OUTPUT_DIR = "latency-stats"
CACHE_OUTPUT_DIR = "cache-stats"
# --- EXECUTION LOGIC ---

def run_single_benchmark(config_id, run_num, benchmark_file, old_freq, new_freq, old_cpuset, new_cpuset):
    """Constructs and executes the Go command for a single test run."""
    
    # Generate unique names for output and cgroup
    timestamp = int(time.time())
    
    # Naming convention: {benchmark_name}_{config_id}_{run_num}.csv
    benchmark_name = os.path.basename(benchmark_file).split('.')[0]
    latency_output_filename = os.path.join(
        LATENCY_OUTPUT_DIR, 
        f"{benchmark_name}_f{old_freq}_{new_freq}_latency.csv"
    )

    cache_stats_output_filename = os.path.join(
        CACHE_OUTPUT_DIR,
        f"{benchmark_name}_f{old_freq}_{new_freq}.csv"
    )


    # Use a unique cgroup name for each run
    cgroup_name = f"{BASE_CGROUP_NAME}"

    # Construct the command list for subprocess
    command = [
        GO_BENCHMARK_RUNNER,
        f"-cgroup-name={cgroup_name}",
        f"-cpuset={old_cpuset}",
        f"-memory={CGROUP_MEMORY_LIMIT}",
        f"-benchmark-file={benchmark_file}",
        f"-new-cpuset={new_cpuset}",
        f"-latency-output-file={latency_output_filename}",
        f"-cache-stats-output-file={cache_stats_output_filename}"
        f"-old-cpu-freq={old_freq}",
        f"-new-cpu-freq={new_freq}",
    ]

    print(f"\n--- Running Config {config_id} (Run {run_num}/{RUNS_PER_CONFIG}) ---")
    print(f"Command: {' '.join(command)}")
    
    try:
        # Execute the command
        # check=True will raise an exception if the Go program returns a non-zero exit code
        result = subprocess.run(command, check=True, capture_output=True, text=True)
        print(f"Status: SUCCESS. Output saved to {latency_output_filename}")
        print(f"Status: SUCCESS. Output saved to {cache_stats_output_filename}")
        # Optional: Print Go program's output/errors
        if result.stdout:
             print(f"Go Output: {result.stdout.strip()}")

    except subprocess.CalledProcessError as e:
        print(f"Status: FAILED. Go program returned error code {e.returncode}")
        print(f"Go Stderr: {e.stderr.strip()}")
    except FileNotFoundError:
        print(f"Status: FATAL ERROR. Go executable not found at {GO_BENCHMARK_RUNNER}. Have you compiled it?")
    except Exception as e:
        print(f"Status: FATAL ERROR. An unexpected error occurred: {e}")

def main():
    # Ensure the output directory exists
    os.makedirs(LATENCY_OUTPUT_DIR, exist_ok=True)
    os.makedirs(CACHE_OUTPUT_DIR, exist_ok=True)
    # Create the Cartesian product of all configuration variables
    configurations = list(itertools.product(
        BENCHMARK_FILES, 
        FREQUENCY_PAIRS, 
        CORE_MIGRATIONS
    ))
    
    total_configs = len(configurations)
    total_runs = total_configs * RUNS_PER_CONFIG
    
    print(f"Orchestrator starting {total_runs} total runs across {total_configs} unique configurations.")

    config_id = 0
    for benchmark_file, freq_pair, core_pair in configurations:
        config_id += 1
        old_freq, new_freq = freq_pair
        old_cpuset, new_cpuset = core_pair
        
        # Loop for the required number of repetitions
        for run_num in range(1, RUNS_PER_CONFIG + 1):
            run_single_benchmark(
                config_id, 
                run_num, 
                benchmark_file, 
                old_freq, 
                new_freq, 
                old_cpuset, 
                new_cpuset
            )
            # Optional: Add a small delay between runs to let the system cool down/settle
            # time.sleep(1)

    print("\n\n--- ALL BENCHMARK RUNS COMPLETE ---")

if __name__ == "__main__":
    main()