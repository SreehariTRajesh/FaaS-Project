import subprocess
import os
import sys

# --- Configuration ---
# The base command to be executed


# The number of times to run the command
NUM_RUNS = 100

# --- Helper Function to Execute Command ---
benchmarks = [
    "cnnserv",
    "imagepr",
    "linpack",
    "lrserv",
    "mlinf",
    "rnnserv",
    "thumbnail",
    "videoproc",
    "vidpr",
    "webserv",
    "wordcnt",
    "compression",
    "graphproc",
    "thumbnail",
]

"""

"""

frequencies = [
    1.0,
    1.25,
    1.50,
    1.75,
    2.0,
    2.25,
    2.50,
    2.75,
    3.00,
    3.25,
    3.50,
    3.75,
    4.00,
    4.25,
    4.50,
    4.75,
    5.00,
]


def run_command(bench, run_number, freqKHz):
    BASE_COMMAND = [
        f"sudo",
        f"./bin/cli",
        f"-cgroup-name={bench}-bench",
        f"-curr-cpuset",
        "0",
        f"-curr-cpu-freq",
        f"{int(freqKHz*1000000)}",
        f"-memory",
        "256M",
        f"-benchmark-file",
        f"serverless-benchmarks/{bench}.py",
        f"-proc-output-file",
        f"process/{bench}_{freqKHz}khz.csv",
        f"-hardware-output-file",
        f"hardware/{bench}_{freqKHz}khz.csv",
    ]
    """
    Executes the command and handles the output.
    """
    print(f"--- Starting Run {run_number}/{NUM_RUNS} ---")

    # We use subprocess.run for simplicity, capturing output and checking for errors
    try:
        # NOTE: We set check=True to raise an error if the command fails (returns non-zero exit code).
        # shell=False is generally safer as it avoids shell injection issues.
        result = subprocess.run(
            BASE_COMMAND, check=True, text=True, capture_output=True
        )

        print(f"Run {run_number} successful.")
        # You can uncomment the lines below if you want to see the command output for debugging
        # print("Stdout:")
        # print(result.stdout)
        # print("Stderr:")
        # print(result.stderr)

    except subprocess.CalledProcessError as e:
        # This block catches errors if the command returns a non-zero exit code
        print(
            f"\n!!! ERROR: Run {run_number} failed with exit code {e.returncode} !!!",
            file=sys.stderr,
        )
        print("Stderr Output:", e.stderr, file=sys.stderr)
        # Decide whether to continue or stop on error
        # sys.exit(1) # Uncomment this to stop the script immediately on the first error
    except FileNotFoundError:
        print(
            f"\n!!! ERROR: Command or script not found. Make sure './bin/cli' exists and is executable.",
            file=sys.stderr,
        )
        sys.exit(1)
    except Exception as e:
        print(
            f"\n!!! An unexpected error occurred during Run {run_number}: {e} !!!",
            file=sys.stderr,
        )
        # sys.exit(1) # Uncomment this to stop the script immediately on the first unexpected error


# --- Main Execution Loop ---


def main():

    for bench in benchmarks:
        for freq in frequencies:
            for i in range(1, NUM_RUNS + 1):
                run_command(bench=bench, run_number=i, freqKHz=freq)
    print("-" * 50)
    print(f"Execution complete. Command was run {NUM_RUNS} times.")


if __name__ == "__main__":
    # Ensure the script is run from the directory containing ./bin/cli if paths are relative
    # print(f"Current working directory: {os.getcwd()}")
    main()
