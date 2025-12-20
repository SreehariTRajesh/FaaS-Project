import time
import numpy as np
import json
import hashlib
import base64
from datetime import datetime
from typing import Dict, Any
import sys

# ============================================================================
# 1. WebSrv - Web Service (JSON Processing + String Operations)
# ============================================================================
class WebSrv:
    """Simulates web request processing with JSON parsing and response generation"""
    
    def __init__(self):
        self.request_count = 0
    
    def process_request(self, payload_size: int = 1000) -> Dict[str, Any]:
        """Process a simulated HTTP request"""
        self.request_count += 1
        
        # Simulate request parsing
        request_data = {
            "user_id": hashlib.md5(str(self.request_count).encode()).hexdigest(),
            "timestamp": datetime.now().isoformat(),
            "payload": "x" * payload_size,
            "headers": {f"header_{i}": f"value_{i}" for i in range(20)}
        }
        
        # JSON serialization/deserialization
        json_str = json.dumps(request_data)
        parsed = json.loads(json_str)
        
        # String processing
        processed = parsed["payload"].upper()[:100]
        hash_val = hashlib.sha256(processed.encode()).hexdigest()
        
        return {
            "status": 200,
            "request_id": self.request_count,
            "hash": hash_val,
            "processed_length": len(processed)
        }
    
    def run_continuous(self, duration: float = 60.0):
        """Run continuously for specified duration"""
        print(f"[WebSrv] Starting continuous execution for {duration}s")
        start_time = time.time()
        iterations = 0
        
        while time.time() - start_time < duration:
            result = self.process_request(payload_size=1000)
            iterations += 1
            
            if iterations % 100 == 0:
                elapsed = time.time() - start_time
                print(f"[WebSrv] Processed {iterations} requests in {elapsed:.2f}s "
                      f"({iterations/elapsed:.2f} req/s)")
        
        total_time = time.time() - start_time
        print(f"[WebSrv] Completed {iterations} iterations in {total_time:.2f}s")
        return iterations
    
if __name__ == '__main__':
    serv = WebSrv()
    serv.run_continuous(duration=60)