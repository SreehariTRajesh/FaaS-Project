import numpy as np 
import time 
import gzip
import json

class CompressionBenchmark:
    """Benchmark for data compression operations"""
    
    def __init__(self, data_size_mb=10):
        self.data_size = data_size_mb * 1024 * 1024
        self.text_data = None
        self.binary_data = None
        self.json_data = None
        
    def setup(self):
        """Generate test data"""
        print(f"Generating {self.data_size // (1024*1024)}MB of test data...")
        
        # Text data (highly compressible)
        self.text_data = b"Lorem ipsum dolor sit amet, consectetur adipiscing elit. " * (self.data_size // 100)
        self.text_data = self.text_data[:self.data_size]
        
        # Binary data (less compressible)
        self.binary_data = np.random.randint(0, 256, self.data_size, dtype=np.uint8).tobytes()
        
        # JSON data (structured)
        records = []
        for i in range(self.data_size // 200):
            records.append({
                'id': i,
                'name': f'user_{i}',
                'value': float(i * 3.14),
                'tags': ['tag1', 'tag2', 'tag3']
            })
        self.json_data = json.dumps(records).encode('utf-8')
    
    def run(self):
        """Run compression benchmarks"""
        # Test 1: Gzip compression of text
        start = time.time()
        compressed = gzip.compress(self.text_data, compresslevel=6)
        elapsed = time.time() - start
        ratio = len(self.text_data) / len(compressed)
        throughput = (len(self.text_data) / (1024*1024)) / elapsed
        self.results['Gzip text (level 6)'] = f"{elapsed:.3f}s, {throughput:.1f} MB/s, ratio: {ratio:.1f}x"
        
        # Test 2: Gzip compression of binary
        start = time.time()
        compressed = gzip.compress(self.binary_data, compresslevel=6)
        elapsed = time.time() - start
        ratio = len(self.binary_data) / len(compressed)
        throughput = (len(self.binary_data) / (1024*1024)) / elapsed
        self.results['Gzip binary (level 6)'] = f"{elapsed:.3f}s, {throughput:.1f} MB/s, ratio: {ratio:.1f}x"
        
        # Test 3: Fast compression
        start = time.time()
        compressed = gzip.compress(self.text_data, compresslevel=1)
        elapsed = time.time() - start
        ratio = len(self.text_data) / len(compressed)
        throughput = (len(self.text_data) / (1024*1024)) / elapsed
        self.results['Gzip text (level 1, fast)'] = f"{elapsed:.3f}s, {throughput:.1f} MB/s, ratio: {ratio:.1f}x"
        
        # Test 4: Max compression
        start = time.time()
        compressed = gzip.compress(self.text_data, compresslevel=9)
        elapsed = time.time() - start
        ratio = len(self.text_data) / len(compressed)
        throughput = (len(self.text_data) / (1024*1024)) / elapsed
        self.results['Gzip text (level 9, max)'] = f"{elapsed:.3f}s, {throughput:.1f} MB/s, ratio: {ratio:.1f}x"
        
        # Test 5: Decompression
        compressed = gzip.compress(self.text_data, compresslevel=6)
        start = time.time()
        decompressed = gzip.decompress(compressed)
        elapsed = time.time() - start
        throughput = (len(decompressed) / (1024*1024)) / elapsed
        self.results['Gzip decompression'] = f"{elapsed:.3f}s, {throughput:.1f} MB/s"

if __name__ == '__main__':
    compression = CompressionBenchmark()
    compression.setup()
    compression.run()
    