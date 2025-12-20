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
# 3. RnnSrv - RNN Service (Sequential Processing)
# ============================================================================
class RnnSrv:
    """Simulates RNN inference with sequential matrix operations"""
    
    def __init__(self, seq_length: int = 100, hidden_size: int = 128):
        self.seq_length = seq_length
        self.hidden_size = hidden_size
        self.vocab_size = 1000
        self.inference_count = 0
        
        # Initialize weights (simulated)
        self.Wxh = np.random.randn(hidden_size, self.vocab_size) * 0.01
        self.Whh = np.random.randn(hidden_size, hidden_size) * 0.01
        self.Why = np.random.randn(self.vocab_size, hidden_size) * 0.01
    
    def forward_pass(self, inputs: np.ndarray) -> np.ndarray:
        """Simulate RNN forward pass"""
        h = np.zeros((self.hidden_size, 1))
        outputs = []
        
        for x in inputs:
            # RNN cell computation
            h = np.tanh(np.dot(self.Wxh, x) + np.dot(self.Whh, h))
            y = np.dot(self.Why, h)
            outputs.append(y)
        
        return np.array(outputs)
    
    def inference(self) -> Dict[str, Any]:
        """Run RNN inference"""
        self.inference_count += 1
        
        # Generate input sequence (one-hot encoded)
        inputs = []
        for _ in range(self.seq_length):
            x = np.zeros((self.vocab_size, 1))
            x[np.random.randint(0, self.vocab_size)] = 1
            inputs.append(x)
        
        # Forward pass
        outputs = self.forward_pass(inputs)
        
        # Get predictions
        predictions = np.argmax(outputs, axis=1)
        
        return {
            "inference_id": self.inference_count,
            "sequence_length": self.seq_length,
            "output_shape": outputs.shape,
            "sample_predictions": predictions[:5].tolist()
        }
    
    def run_continuous(self, duration: float = 60.0):
        """Run continuously for specified duration"""
        print(f"[RnnSrv] Starting continuous execution for {duration}s")
        start_time = time.time()
        iterations = 0
        
        while time.time() - start_time < duration:
            result = self.inference()
            iterations += 1
            
            if iterations % 5 == 0:
                elapsed = time.time() - start_time
                print(f"[RnnSrv] Completed {iterations} inferences in {elapsed:.2f}s "
                      f"({iterations/elapsed:.2f} inf/s)")
        
        total_time = time.time() - start_time
        print(f"[RnnSrv] Completed {iterations} iterations in {total_time:.2f}s")
        return iterations
    
if __name__ == '__main__':
    rnnserv = RnnSrv(seq_length=200, hidden_size=128)
    rnnserv.run_continuous(duration=60)
    