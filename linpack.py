import numpy as np
import time

def linpack_benchmark(n=2000):
    """
    Solves a dense system of linear equations Ax = b.
    n: The size of the matrix (n x n). 
       n=2000 is a standard 'heavy' load for modern CPUs.
    """
    print(f"Running Linpack Benchmark (Matrix Size: {n}x{n})...")

    # 1. Data Generation (Memory heavy)
    # We use float64 (double precision) which is standard for Linpack
    A = np.random.rand(n, n).astype(np.float64)
    b = np.random.rand(n).astype(np.float64)

    # 2. The Benchmark (Compute heavy)
    start_time = time.time_ns()
    
    # solving Ax = b uses LU decomposition (the core of Linpack)
    x = np.linalg.solve(A, b)
    
    end_time = time.time_ns()
    
    # 3. Validation (Optional but recommended)
    # Check if Ax - b is close to 0
    # norm = np.linalg.norm(np.dot(A, x) - b)

    # 4. Calculate Performance
    duration = end_time - start_time
    
    # The FLOP count for solving a linear system is approx (2/3) * n^3
    ops = (2.0/3.0) * (n ** 3)
    
    # Giga-FLOPS = (Operations / 10^9) / Time
    gflops = (ops * 1e-9) / duration

    print(f"Done.")
    print(f"Duration: {duration:.4f} nano seconds")
    print(f"Performance: {gflops:.4f} GFLOPS")
    
    return gflops

if __name__ == "__main__":
    # Run a warm-up to load libraries into RAM
    
    # Run the real test
    linpack_benchmark(n=10)