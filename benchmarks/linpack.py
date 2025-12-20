import time
import numpy as np
import json
import hashlib
import base64
from datetime import datetime
from typing import Dict, Any
import sys

class Linpack:
    """LINPACK-style benchmark for linear algebra operations"""
    
    def __init__(self, matrix_size: int = 500):
        self.matrix_size = matrix_size
        self.benchmark_count = 0
    
    def run_benchmark(self) -> Dict[str, Any]:
        """Run LINPACK-style benchmark"""
        self.benchmark_count += 1
        
        # Generate random matrix and vector
        A = np.random.randn(self.matrix_size, self.matrix_size)
        b = np.random.randn(self.matrix_size)
        
        # Time the solve operation
        start = time.time()
        
        # LU decomposition and solve
        x = np.linalg.solve(A, b)
        
        solve_time = time.time() - start
        
        # Verify solution
        residual = np.linalg.norm(np.dot(A, x) - b)
        
        # Additional operations
        # Matrix multiplication
        mm_start = time.time()
        C = np.dot(A, A.T)
        mm_time = time.time() - mm_start
        
        # Eigenvalue computation (subset)
        eigen_start = time.time()
        eigenvalues = np.linalg.eigvalsh(A[:100, :100])
        eigen_time = time.time() - eigen_start
        
        # Calculate FLOPS estimate
        # Solve: ~(2/3) * n^3 operations
        flops_solve = (2.0 / 3.0) * (self.matrix_size ** 3)
        gflops_solve = (flops_solve / solve_time) / 1e9
        
        return {
            "benchmark_id": self.benchmark_count,
            "matrix_size": self.matrix_size,
            "solve_time": solve_time,
            "mm_time": mm_time,
            "eigen_time": eigen_time,
            "gflops": gflops_solve,
            "residual": float(residual)
        }
    
    def run_continuous(self, duration: float = 60.0):
        """Run continuously for specified duration"""
        print(f"[Linpack] Starting continuous execution for {duration}s")
        start_time = time.time()
        iterations = 0
        total_gflops = 0.0
        
        while time.time() - start_time < duration:
            result = self.run_benchmark()
            iterations += 1
            total_gflops += result["gflops"]
            
            if iterations % 5 == 0:
                elapsed = time.time() - start_time
                avg_gflops = total_gflops / iterations
                print(f"[Linpack] Completed {iterations} benchmarks in {elapsed:.2f}s "
                      f"(Avg: {avg_gflops:.2f} GFLOPS)")
        
        total_time = time.time() - start_time
        avg_gflops = total_gflops / iterations
        print(f"[Linpack] Completed {iterations} iterations in {total_time:.2f}s "
              f"(Average: {avg_gflops:.2f} GFLOPS)")
        return iterations
    
if __name__ == '__main__':
    linpk = Linpack(matrix_size=500)
    linpk.run_continuous(duration=60)