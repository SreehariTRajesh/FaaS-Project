#!/bin/bash

# CPUs to isolate
ISOLATED_CPUS="0-3"
SYSTEM_CPUS="4-31"

echo "Isolating CPUs $ISOLATED_CPUS..."

# 1. Move ALL processes to system CPUs
for PID in $(ps -eLo pid | tail -n +2); do
    taskset -acp $SYSTEM_CPUS $PID 2>/dev/null
done

# 2. Move ALL kernel threads
for PID in $(ps -eLo pid,comm | grep '\[.*\]' | awk '{print $1}'); do
    taskset -acp $SYSTEM_CPUS $PID 2>/dev/null
done

# 3. Explicitly move common kernel threads
for KTHREAD in kworker ksoftirqd migration rcu watchdog; do
    for PID in $(pgrep -f "^\[$KTHREAD"); do
        taskset -acp $SYSTEM_CPUS $PID 2>/dev/null
    done
done

# 4. Set default IRQ affinity
echo $SYSTEM_CPUS > /proc/irq/default_smp_affinity

# 5. Move ALL existing IRQs
for IRQ in /proc/irq/*/smp_affinity_list; do
    echo $SYSTEM_CPUS > $IRQ 2>/dev/null
done

# 6. Disable specific per-CPU kernel threads on isolated CPUs
for CPU in 2 3; do
    # Try to disable watchdog
    echo 0 > /sys/devices/system/cpu/cpu$CPU/online 2>/dev/null
    echo 1 > /sys/devices/system/cpu/cpu$CPU/online 2>/dev/null
done

echo "CPU isolation complete!"
echo "Use: taskset -c $ISOLATED_CPUS ./your_program"
