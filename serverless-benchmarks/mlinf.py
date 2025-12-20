import numpy as np
import time 

class MLInferenceBenchmark:
    """Benchmark for ML inference operations"""
    
    def __init__(self, batch_size=32, n_batches=100):
        self.batch_size = batch_size
        self.n_batches = n_batches
        self.models = {}
        self.data = None
        
    def setup(self):
        """Setup simple ML models and data"""
        print(f"Setting up ML models and {self.n_batches} batches of size {self.batch_size}...")
        
        # Generate synthetic input data (image-like)
        self.data = [
            np.random.randn(self.batch_size, 224, 224, 3).astype(np.float32)
            for _ in range(self.n_batches)
        ]
        
        # Simple linear model (matrix multiplication)
        self.models['linear'] = {
            'weights': np.random.randn(224*224*3, 1000).astype(np.float32),
            'bias': np.random.randn(1000).astype(np.float32)
        }
        
        # Simple conv-like operation weights
        self.models['conv'] = {
            'kernel': np.random.randn(3, 3, 3, 64).astype(np.float32)
        }
    
    def run(self):
        """Run ML inference benchmarks"""
        # Test 1: Linear layer inference
        start = time.time()
        for batch in self.data:
            flat = batch.reshape(self.batch_size, -1)
            output = np.dot(flat, self.models['linear']['weights']) + self.models['linear']['bias']
        elapsed = time.time() - start
        throughput = (self.n_batches * self.batch_size) / elapsed
        self.results['Linear layer'] = f"{elapsed:.3f}s ({throughput:.1f} samples/sec)"
        
        # Test 2: ReLU activation
        start = time.time()
        for batch in self.data:
            output = np.maximum(0, batch)
        elapsed = time.time() - start
        throughput = (self.n_batches * self.batch_size) / elapsed
        self.results['ReLU activation'] = f"{elapsed:.3f}s ({throughput:.1f} samples/sec)"
        
        # Test 3: Softmax
        start = time.time()
        for batch in self.data[:20]:  # Subset for expensive operation
            flat = batch.reshape(self.batch_size, -1)
            exp_x = np.exp(flat - np.max(flat, axis=1, keepdims=True))
            output = exp_x / np.sum(exp_x, axis=1, keepdims=True)
        elapsed = time.time() - start
        throughput = (20 * self.batch_size) / elapsed
        self.results['Softmax (sample)'] = f"{elapsed:.3f}s ({throughput:.1f} samples/sec)"
        
        # Test 4: Batch normalization
        start = time.time()
        for batch in self.data:
            mean = np.mean(batch, axis=0, keepdims=True)
            var = np.var(batch, axis=0, keepdims=True)
            output = (batch - mean) / np.sqrt(var + 1e-5)
        elapsed = time.time() - start
        throughput = (self.n_batches * self.batch_size) / elapsed
        self.results['Batch normalization'] = f"{elapsed:.3f}s ({throughput:.1f} samples/sec)"
        
        # Test 5: Dropout (inference mode - just pass through with mask)
        start = time.time()
        for batch in self.data:
            mask = np.random.binomial(1, 0.8, batch.shape)
            output = batch * mask
        elapsed = time.time() - start
        throughput = (self.n_batches * self.batch_size) / elapsed
        self.results['Dropout application'] = f"{elapsed:.3f}s ({throughput:.1f} samples/sec)"


if __name__ == '__main__':
    mlinf = MLInferenceBenchmark()
    mlinf.setup()
    mlinf.run()
    