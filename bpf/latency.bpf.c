// go:build ignore

#include "headers/vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

// ... (struct migration_event_t and migration_map remain the same) ...

struct migration_event_t
{
    __u64 timestamp_start;
    __u64 timestamp_end;
    __u64 latency;
    __u32 pid;
    __u32 source_cpu;
    __u32 target_cpu;
    __u32 cgroup_id;
};

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, __u32);
    __type(value, bool);
} container_map SEC(".maps");

// Use map type BPF_MAP_TYPE_HASH
struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, __u64);
    __type(value, struct migration_event_t);
} migration_map SEC(".maps");

// Optional: Ringbuffer to send completed events to userspace
struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1024 * 1024);
} events SEC(".maps");

SEC("raw_tracepoint/sched_migrate_task")
int tracepoint_sched_migrate_task(struct bpf_raw_tracepoint_args *ctx)
{
    __u32 pid = 0;
    struct task_struct *p = (struct task_struct *)ctx->args[0];
    __u32 dest_cpu = (__u32)ctx->args[1];
    struct cgroup *cg;
    struct kernfs_node *kn;
    struct css_set *cgroups;
    u64 cgroup_id = 0;
    __u32 curr_cpu = 0;
    bpf_probe_read_kernel(&curr_cpu, sizeof(__u32), &p->thread_info.cpu);

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

    if (p)
    {
        bpf_probe_read_kernel(&pid, sizeof(__u32), &p->pid);
    }

    if (!bpf_map_lookup_elem(&container_map, &cgroup_id))
    {
        return 0;
    }

    struct migration_event_t event = {};

    event.timestamp_start = bpf_ktime_get_ns();
    event.cgroup_id = cgroup_id;
    event.pid = pid;

    // FIX 2: Use BPF_CORE_READ_INTO for stable access to tracepoint args
    bpf_probe_read_kernel(&event.source_cpu, sizeof(__u32), &p->thread_info.cpu);
    bpf_probe_read_kernel(&event.target_cpu, sizeof(__u32), &dest_cpu);

    __u64 key = (__u64)cgroup_id << 32 | (__u64)pid;

    // FIX 3: Use BPF_ANY to ensure the tracking starts
    bpf_map_update_elem(&migration_map, &key, &event, BPF_ANY);

    return 0;
}
// --- Program 2: End Tracking on Task Start (Switch-in) ---
// Using raw_tracepoint is often preferred for sched_switch as it provides direct context.
SEC("raw_tracepoint/sched_switch")
int raw_tp_sched_switch(struct bpf_raw_tracepoint_args *ctx)
{
    __u32 next_pid = 0;
    struct task_struct *p = (struct task_struct *)ctx->args[1];
    struct cgroup *cg;
    struct kernfs_node *kn;
    struct css_set *cgroups;
    __u64 cgroup_id = 0;

    // First, safely read the cgroups pointer from task->cgroups
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

    if (p)
    {
        bpf_probe_read_kernel(&next_pid, sizeof(__u32), &p->pid);
    }

    __u64 end_time = bpf_ktime_get_ns();

    __u64 key = (__u64)cgroup_id << 32 | (__u64)next_pid;

    // Check if we are tracking this task's migration
    struct migration_event_t *event = bpf_map_lookup_elem(&migration_map, &key);

    if (event)
    {
        // 1. Calculate latency
        event->timestamp_end = end_time;
        event->latency = event->timestamp_end - event->timestamp_start;

        // 2. Print output using correct format specifiers (%llu for __u64)
        bpf_printk("Migration Latency: %llu ns, PID:%d, CGroup:%llu, CPU:%d -> %d",
                   event->latency,
                   event->pid,
                   event->cgroup_id,
                   event->source_cpu,
                   event->target_cpu);

        struct migration_event_t *rb_event;
        rb_event = bpf_ringbuf_reserve(&events, sizeof(*rb_event), 0);

        if (rb_event)
        {
            __builtin_memcpy(rb_event, event, sizeof(*rb_event));
            bpf_ringbuf_submit(rb_event, 0);
        }
        // 3. Delete the element to finalize measurement and clean up the map
        bpf_map_delete_elem(&migration_map, &key);
    }
    return 0;
}

char LICENSE[] SEC("license") = "GPL";