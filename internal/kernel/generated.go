package kernel

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go latency ../../bpf/latency.bpf.c -- -I../../bpf/headers
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go llc ../../bpf/llc.bpf.c -- -I../../bpf/headers
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go proc ../../bpf/proc.bpf.c -- -I../../bpf/headers
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go hardware ../../bpf/hardware.bpf.c -- -I../../bpf/headers
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go function ../../bpf/function.bpf.c -- -I../../bpf/headers
