package main

import (
	"faas-migration/internal/benchmarks"
	"flag"
	"log"
)

func main() {
	cgroupPtr := flag.String("cgroup-name", "benchmark-cgroup", "Name of the cgroup to manage (e.g., 'migration_test').")
	cpusetPtr := flag.String("curr-cpuset", "0", "CPU set for the cgroup (e.g., '0-3').")
	currCPUFreqPtr := flag.Uint64("curr-cpu-freq", 1000000, "Old CPU frequency for migration (e.g., 1000000).")
	memoryPtr := flag.String("memory", "512M", "Memory limit for the cgroup (e.g., '512M').")
	benchmarkFile := flag.String("benchmark-file", "", "Path to the benchmark")
	procOutputFilePtr := flag.String("proc-output-file", "proc-bench.csv", "Output file for proc metrics.")
	hardwareOutputFilePtr := flag.String("hardware-output-file", "hw-bench.csv", "Output file for hardware metrics.")
	flag.Parse()

	runner, err := benchmarks.NewProcBenchmarkRunner(*cgroupPtr, *procOutputFilePtr, *hardwareOutputFilePtr)

	if err != nil {
		log.Fatalf("error initializing proc benchmark runner: %v", err)
	}

	err = runner.RunProcBenchmark(*benchmarkFile, *cpusetPtr, *memoryPtr, *currCPUFreqPtr)

	/*
			runner, err := container.NewBenchmarkRunner(*cgroupPtr, *latencyOutputFilePtr, *cacheStatsOutputFilePtr)
		if err != nil {
			log.Fatalf("Error initializing benchmark runner: %v", err)
		}

		err = runner.RunBenchmark(*benchmarkFile, *cpusetPtr, *memoryPtr, *newCPUSetPtr, *oldCPUFreqPtr, *newCPUFreqPtr)
		if err != nil {
			log.Fatalf("Error running benchmark: %v", err)
		}

			runner, err := benchmarks.NewHardwarBenchmarkRunner("hw.csv")

			if err != nil {
				log.Fatalln("Error initializing hardware benchmark: %w", err)
			}

			err = runner.RunBenchmark()
			if err != nil {
				log.Fatalln("Error while running hardware benchmark: %w", err)
			}
	*/
	if err != nil {
		log.Fatalf("error running proc benchmark: %v", err)
	}
}
