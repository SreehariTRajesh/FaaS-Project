#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>


struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1024 * 1024);
} function_events SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, __u64); // Use u64 to capture full PID+TID
    __type(value, __u64);
} start_times SEC(".maps");

struct run_event
{
    __u32 pid;
    __u64 duration_ns;
};


SEC("uprobe")
int uprobe_entry(struct pt_regs *ctx)
{
    __u64 id = bpf_get_current_pid_tgid();
    __u64 start_ts = bpf_ktime_get_ns();

    bpf_map_update_elem(&start_times, &id, &start_ts, BPF_ANY);
    return 0;
}

SEC("uretprobe")
int uprobe_exit(struct pt_regs *ctx)
{
    __u64 id = bpf_get_current_pid_tgid();
    __u32 pid = id >> 32;

    __u64 *start_ts = bpf_map_lookup_elem(&start_times, &id);
    if (!start_ts)
    {
        // If this happens, the entry probe didn't fire or map was full
        return 0;
    }

    __u64 end_ts = bpf_ktime_get_ns();
    __u64 duration = end_ts - *start_ts;

    struct run_event *e = bpf_ringbuf_reserve(&function_events, sizeof(*e), 0);
    if (e)
    {
        e->duration_ns = duration;
        e->pid = pid;
        bpf_ringbuf_submit(e, 0);
    }

    bpf_map_delete_elem(&start_times, &id);
    return 0;
}

char LICENSE[] SEC("license") = "GPL";