// go:build ignore

#include "headers/vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

struct proc_event_t
{
    __u32 pid;
    __u32 cgroup_id;
    __u64 start_timestamp;
    __u64 end_timestamp;
    __u64 latency;
};

// Optional: Ringbuffer to send completed events to userspace
struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} events SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 16);
    __type(key, __u32);
    __type(value, bool);
} process_container_map SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 16);
    __type(key, __u32);
    __type(value, struct proc_event_t);
} process_monitor_map SEC(".maps");


SEC("tracepoint/syscalls/sys_enter_execve")
int trace_execve(struct trace_event_raw_sys_enter *ctx)
{
    struct proc_event_t event = {};
    struct task_struct *p;
    struct cgroup *cg;
    struct kernfs_node *kn;
    struct css_set *cgroups;
    __u64 cgroup_id = 0;

    p = (struct task_struct *)bpf_get_current_task();
    __u32 pid = bpf_get_current_pid_tgid() >> 32;

    bpf_probe_read_kernel(&cgroups, sizeof(cgroups), &p->cgroups);
    if (cgroups)
    {
        // Then read the cgroup pointer from cgroups->dfl_cgrp
        bpf_probe_read_kernel(&cg, sizeof(cg), &cgroups->dfl_cgrp);
        if (cg)
        {
            // Read the kernfs_node pointer from cgroup->kn
            bpf_probe_read_kernel(&kn, sizeof(kn), &cg->kn);
            if (kn)
            {
                // Read the id from kernfs_node->id
                bpf_probe_read_kernel(&cgroup_id, sizeof(cgroup_id), &kn->id);
            }
        }
    }
    
    bpf_printk("PID: %d CGROUP ID: %d", pid, cgroup_id);

    if (!bpf_map_lookup_elem(&process_container_map, &cgroup_id))
        return 0;

    event.start_timestamp = bpf_ktime_get_ns();
    event.pid = pid;
    event.cgroup_id = cgroup_id;

    bpf_map_update_elem(&process_monitor_map, &pid, &event, BPF_NOEXIST);

    return 0;
}

SEC("kprobe/do_exit")
int kprobe_do_exit(struct pt_regs *ctx)
{
    struct task_struct *p;
    p = (struct task_struct *)bpf_get_current_task();
    __u32 pid = bpf_get_current_pid_tgid() >> 32;
    __u64 end_timestamp = bpf_ktime_get_ns();
    struct proc_event_t *event = bpf_map_lookup_elem(&process_monitor_map, &pid);
    if (event)
    {
        event->end_timestamp = end_timestamp;
        event->latency = event->end_timestamp - event->start_timestamp;

        struct proc_event_t *rb_event;
        rb_event = bpf_ringbuf_reserve(&events, sizeof(*rb_event), 0);

        if (rb_event)
        {
            __builtin_memcpy(rb_event, event, sizeof(*rb_event));
            bpf_ringbuf_submit(rb_event, 0);
        }
    }
    return 0;
}


char _license[] SEC("license") = "GPL";