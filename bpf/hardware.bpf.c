#include <vmlinux.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <stdbool.h>

struct perf_stats
{
    __u64 cycles;
    __u64 instructions;
    __u64 ref_cycles;
    __u64 cache_references;
    __u64 cache_misses;
    __u64 branches;
    __u64 branch_misses;
    __u64 l1d_loads;
    __u64 l1d_load_misses;
    __u64 l1d_stores;
    __u64 l1d_store_misses;
    __u64 l1d_prefetches;
    __u64 l1d_prefetch_misses;
    __u64 l1i_loads;
    __u64 l1i_load_misses;
    __u64 l1i_prefetches;
    __u64 llc_loads;
    __u64 llc_load_misses;
    __u64 llc_stores;
    __u64 llc_store_misses;
    __u64 llc_prefetches;
    __u64 llc_prefetch_misses;
    __u64 dtlb_loads;
    __u64 dtlb_load_misses;
    __u64 dtlb_stores;
    __u64 dtlb_store_misses;
    __u64 dtlb_prefetches;
    __u64 dtlb_prefetch_misses;
    __u64 tlb_loads;
    __u64 tlb_load_misses;
    __u64 bpu_loads;
    __u64 bpu_load_misses;
};

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct perf_stats);
} stats SEC(".maps");

SEC("perf_event")
int on_cpu_cycles(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->cycles, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_instructions(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->instructions, ctx->sample_period);

    return 0;
}

SEC("perf_event")
int on_ref_cycles(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->ref_cycles, ctx->sample_period);

    return 0;
}

SEC("perf_event")
int on_cache_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->cache_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_cache_references(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->cache_references, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_branches(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->branches, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_branch_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->branch_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1d_loads(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1d_loads, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1d_load_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1d_load_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1d_stores(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1d_stores, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1d_store_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1d_stores, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1d_prefetches(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1d_prefetches, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1d_prefetch_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1d_prefetch_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1i_loads(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1i_loads, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1i_load_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1i_load_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_l1i_prefetches(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->l1i_prefetches, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_llc_loads(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->llc_loads, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_llc_load_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->llc_load_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_llc_stores(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->llc_stores, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_llc_store_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->llc_store_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_llc_prefetches(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->llc_prefetches, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_llc_prefetch_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->llc_prefetch_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_dtlb_loads(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->dtlb_loads, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_dtlb_load_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->dtlb_load_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_dtlb_stores(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->dtlb_stores, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_dtlb_store_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->dtlb_store_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_dtlb_prefetches(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->dtlb_prefetches, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_dtlb_prefetch_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->dtlb_prefetch_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_tlb_loads(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->tlb_loads, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_tlb_load_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->tlb_load_misses, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_bpu_loads(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->bpu_loads, ctx->sample_period);
    return 0;
}

SEC("perf_event")
int on_bpu_load_misses(struct bpf_perf_event_data *ctx)
{
    __u32 stats_key = (__u32)bpf_get_current_cgroup_id();
    struct perf_stats *s = bpf_map_lookup_elem(&stats, &stats_key);
    if (!s)
    {
        return 0;
    }

    __sync_fetch_and_add(&s->bpu_load_misses, ctx->sample_period);
    return 0;
}

char LICENSE[] SEC("license") = "GPL";