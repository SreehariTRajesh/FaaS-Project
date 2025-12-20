"""
Simulated Serverless Benchmarks for Continuous Running Functions
Designed for CPU migration experiments in containerized environments
"""

import time
import numpy as np
import json
import hashlib
import base64
from datetime import datetime
from typing import Dict, Any
import sys


# ============================================================================
# 4. CnnSrv - CNN Service (Convolutional Operations)
# ============================================================================
class CnnSrv:
    """Simulates CNN inference with convolutional operations"""
    
    def __init__(self, input_size: tuple = (224, 224, 3)):
        self.input_size = input_size
        self.inference_count = 0
    
    def conv2d(self, input_data: np.ndarray, num_filters: int, kernel_size: int) -> np.ndarray:
        """Simulate 2D convolution"""
        h, w, c = input_data.shape
        output_h = h - kernel_size + 1
        output_w = w - kernel_size + 1
        
        # Generate random filters
        filters = np.random.randn(num_filters, kernel_size, kernel_size, c) * 0.01
        output = np.zeros((output_h, output_w, num_filters))
        
        # Convolution operation
        for f in range(num_filters):
            for i in range(output_h):
                for j in range(output_w):
                    region = input_data[i:i+kernel_size, j:j+kernel_size, :]
                    output[i, j, f] = np.sum(region * filters[f])
        
        return output
    
    def relu(self, x: np.ndarray) -> np.ndarray:
        """ReLU activation"""
        return np.maximum(0, x)
    
    def max_pool(self, input_data: np.ndarray, pool_size: int = 2) -> np.ndarray:
        """Max pooling operation"""
        h, w, c = input_data.shape
        output_h = h // pool_size
        output_w = w // pool_size
        output = np.zeros((output_h, output_w, c))
        
        for i in range(output_h):
            for j in range(output_w):
                for k in range(c):
                    region = input_data[i*pool_size:(i+1)*pool_size, 
                                       j*pool_size:(j+1)*pool_size, k]
                    output[i, j, k] = np.max(region)
        
        return output
    
    def inference(self) -> Dict[str, Any]:
        """Run CNN inference"""
        self.inference_count += 1
        
        # Generate input image
        input_img = np.random.randn(*self.input_size).astype(np.float32)
        
        # Layer 1: Conv + ReLU + MaxPool
        conv1 = self.conv2d(input_img, num_filters=32, kernel_size=3)
        relu1 = self.relu(conv1)
        pool1 = self.max_pool(relu1, pool_size=2)
        
        # Layer 2: Conv + ReLU + MaxPool
        conv2 = self.conv2d(pool1, num_filters=64, kernel_size=3)
        relu2 = self.relu(conv2)
        pool2 = self.max_pool(relu2, pool_size=2)
        
        # Flatten and classify
        flattened = pool2.flatten()
        logits = np.dot(np.random.randn(10, len(flattened)), flattened)
        prediction = np.argmax(logits)
        
        return {
            "inference_id": self.inference_count,
            "input_shape": self.input_size,
            "final_shape": pool2.shape,
            "prediction": int(prediction),
            "confidence": float(np.max(np.exp(logits) / np.sum(np.exp(logits))))
        }
    
    def run_continuous(self, duration: float = 60.0):
        """Run continuously for specified duration"""
        print(f"[CnnSrv] Starting continuous execution for {duration}s")
        start_time = time.time()
        iterations = 0
        
        while time.time() - start_time < duration:
            result = self.inference()
            iterations += 1
            
            if iterations % 5 == 0:
                elapsed = time.time() - start_time
                print(f"[CnnSrv] Completed {iterations} inferences in {elapsed:.2f}s "
                      f"({iterations/elapsed:.2f} inf/s)")
        
        total_time = time.time() - start_time
        print(f"[CnnSrv] Completed {iterations} iterations in {total_time:.2f}s")
        return iterations

if __name__ == '__main__':
    cnnserv = CnnSrv(input_size=(448, 448, 3))
    cnnserv.run_continuous(duration=120)
