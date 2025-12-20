import psutil
import time
import pandas as pd
import matplotlib.pyplot as plt
import threading

# Global flag to stop recording
keep_running = True
data_log = []

def record_frequencies(interval=0.1):
    """Polls CPU frequency every 'interval' seconds."""
    global keep_running
    start_time = time.time()
    
    print("Starting frequency monitoring...")
    while keep_running:
        current_time = time.time() - start_time
        # Get frequency of ALL cores
        freqs = psutil.cpu_freq(percpu=True)
        
        # Extract just the 'current' Mhz value for each core
        row = {'Time': current_time}
        for i, f in enumerate(freqs):
            row[f'Core {i}'] = f.current
            
        data_log.append(row)
        time.sleep(interval)

def visualize_data():
    df = pd.DataFrame(data_log)
    df.set_index('Time', inplace=True)
    
    plt.figure(figsize=(12, 6))
    
    # Plot every core
    for column in df.columns:
        plt.plot(df.index, df[column], label=column, alpha=0.7, linewidth=1.5)

    plt.title("CPU Core Frequencies Over Time")
    plt.ylabel("Frequency (MHz)")
    plt.xlabel("Time (Seconds)")
    plt.legend(loc='upper right', bbox_to_anchor=(1.15, 1))
    plt.grid(True, which='both', linestyle='--', linewidth=0.5)
    plt.tight_layout()
    
    # Save to file
    plt.savefig("cpu_freq_analysis.png")
    print("Graph saved to 'cpu_freq_analysis.png'")
    plt.show()

# --- SIMULATE A WORKLOAD ---
if __name__ == "__main__":
    # 1. Start Recording in a background thread
    monitor_thread = threading.Thread(target=record_frequencies)
    monitor_thread.start()

    # 2. Run your benchmark here (Simulated with a loop)
    print("Running heavy workload...")
    try:
        # Simulate heavy CPU load (Matrix multiplication)
        import numpy as np
        start = time.time()
        while time.time() - start < 5: # Run for 5 seconds
            np.dot(np.random.rand(1000, 1000), np.random.rand(1000, 1000))
    except KeyboardInterrupt:
        pass
    
    # 3. Stop recording and visualize
    print("Workload finished. Generating graph...")
    keep_running = False
    monitor_thread.join()
    visualize_data()