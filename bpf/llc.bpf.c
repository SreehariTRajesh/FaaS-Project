#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <linux/perf_event.h>
#include <stdbool.h>

struct llc_event_t
{
    __u32 pid;
    __u32 cpu;
    __u32 cgroup_id;

    __u64 read_references;
    __u64 read_misses;
    __u64 read_hits;

    __u64 write_references;
    __u64 write_misses;
    __u64 write_hits;

    __u64 prefetch_references;
    __u64 prefetch_misses;
    __u64 prefetch_hits;

    __u64 total_references;
    __u64 total_misses;
    __u64 total_hits;
};

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 16);
    __type(key, __u64);
    __type(value, struct llc_event_t);
} llc_stats_map SEC(".maps");

static __always_inline __u64 __update_stats(__u32 pid, __u32 cpu, bool is_hit, bool is_write, __u8 op_type)
{
    __u64 key = ((__u64)pid << 32) | (__u64)cpu;
    struct llc_event_t *event = bpf_map_lookup_elem(&llc_stats_map, &key);
    if (!event)
    {
        struct llc_event_t new_event = {
            .pid = pid,
            .cpu = cpu,
            .cgroup_id = bpf_get_current_cgroup_id(),
            .read_hits = 0,
            .read_misses = 0,
            .read_references = 0,
            .write_hits = 0,
            .write_misses = 0,
            .write_references = 0,
            .prefetch_hits = 0,
            .prefetch_misses = 0,
            .prefetch_references = 0,
            .total_hits = 0,
            .total_misses = 0,
            .total_references = 0,
        };
        bpf_map_update_elem(&llc_stats_map, &key, &new_event, BPF_ANY);
        event = bpf_map_lookup_elem(&llc_stats_map, &key);
        if(!event) return 0;
    }
    switch (op_type)
    {
    case 0:
        /* read */
        __sync_fetch_and_add(&event->read_references, 1);
        if (is_hit)
        {
            __sync_fetch_and_add(&event->read_hits, 1);
            __sync_fetch_and_add(&event->total_hits, 1);
        }
        else
        {
            __sync_fetch_and_add(&event->read_misses, 1);
            __sync_fetch_and_add(&event->total_misses, 1);
        }
        break;
    case 1:
        /* write */
        __sync_fetch_and_add(&event->write_references, 1);
        if (is_hit)
        {
            __sync_fetch_and_add(&event->write_hits, 1);
            __sync_fetch_and_add(&event->total_hits, 1);
        }
        else
        {
            __sync_fetch_and_add(&event->write_misses, 1);
            __sync_fetch_and_add(&event->total_misses, 1);
        }
        break;
    case 2:
        /* prefetch */
        __sync_fetch_and_add(&event->prefetch_references, 1);
        if (is_hit)
        {
            __sync_fetch_and_add(&event->prefetch_hits, 1);
            __sync_fetch_and_add(&event->total_hits, 1);
        }
        else
        {
            __sync_fetch_and_add(&event->prefetch_misses, 1);
            __sync_fetch_and_add(&event->total_misses, 1);
        }
        break;
    default:
        break;
    }
    return 0;
}


struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 16);
    __type(key, __u32);
    __type(value, bool);
} llc_container_map SEC(".maps");

SEC("perf_event")
int llc_read_miss_handler(struct bpf_perf_event_data *ctx)
{
    __u32 pid = bpf_get_current_pid_tgid() >> 32;
    __u32 cpu = bpf_get_smp_processor_id();
    __u32 cgroup_id = bpf_get_current_cgroup_id();
    if(!bpf_map_lookup_elem(&llc_container_map, &cgroup_id))
        return 0;
    __update_stats(pid, cpu, false, false, 0);
    return 0;
}

SEC("perf_event")
int llc_read_hit_handler(struct bpf_perf_event_data *ctx)
{
    __u32 pid = bpf_get_current_pid_tgid() >> 32;
    __u32 cpu = bpf_get_smp_processor_id();
    __u32 cgroup_id = bpf_get_current_cgroup_id();
    if(!bpf_map_lookup_elem(&llc_container_map, &cgroup_id))
        return 0;
    __update_stats(pid, cpu, true, false, 0);
    return 0;
}

SEC("perf_event")
int llc_write_miss_handler(struct bpf_perf_event_data *ctx)
{
    __u32 pid = bpf_get_current_pid_tgid() >> 32;
    __u32 cpu = bpf_get_smp_processor_id();
    __u32 cgroup_id = bpf_get_current_cgroup_id();
    if(!bpf_map_lookup_elem(&llc_container_map, &cgroup_id))
        return 0;
    __update_stats(pid, cpu, false, true, 1);
    return 0;
}

SEC("perf_event")
int llc_write_hit_handler(struct bpf_perf_event_data *ctx)
{
    __u32 pid = bpf_get_current_pid_tgid() >> 32;
    __u32 cpu = bpf_get_smp_processor_id();
    __u32 cgroup_id = bpf_get_current_cgroup_id();
    if(!bpf_map_lookup_elem(&llc_container_map, &cgroup_id))
        return 0;
    __update_stats(pid, cpu, true, true, 1);
    return 0;
}

SEC("perf_event")
int llc_prefetch_miss_handler(struct bpf_perf_event_data *ctx)
{
    __u32 pid = bpf_get_current_pid_tgid() >> 32;
    __u32 cpu = bpf_get_smp_processor_id();
    __u32 cgroup_id = bpf_get_current_cgroup_id();
    if(!bpf_map_lookup_elem(&llc_container_map, &cgroup_id))
        return 0;
    __update_stats(pid, cpu, false, false, 2);
    return 0;
}

SEC("perf_event")
int llc_prefetch_hit_handler(struct bpf_perf_event_data *ctx)
{
    __u32 pid = bpf_get_current_pid_tgid() >> 32;
    __u32 cpu = bpf_get_smp_processor_id();
    __u32 cgroup_id = bpf_get_current_cgroup_id();
    if(!bpf_map_lookup_elem(&llc_container_map, &cgroup_id))
        return 0;
    __update_stats(pid, cpu, true, false, 2);
    return 0;
}

char _license[] SEC("license") = "GPL";
