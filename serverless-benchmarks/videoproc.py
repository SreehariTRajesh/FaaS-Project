import numpy as np 
import time 
from PIL import Image

class VideoProcessingBenchmark:
    """Benchmark for video processing operations"""
    
    def __init__(self, n_frames=300, resolution=(1280, 720)):
        self.n_frames = n_frames
        self.resolution = resolution
        self.frames = []
        
    def setup(self):
        """Generate synthetic video frames"""
        print(f"Generating {self.n_frames} video frames ({self.resolution[0]}x{self.resolution[1]})...")
        for i in range(self.n_frames):
            frame = np.random.randint(0, 255, 
                                     (self.resolution[1], self.resolution[0], 3), 
                                     dtype=np.uint8)
            self.frames.append(frame)
    
    def run(self):
        """Run video processing benchmarks"""
        # Test 1: Frame differencing (motion detection)
        start = time.time()
        diffs = []
        for i in range(1, len(self.frames)):
            diff = np.abs(self.frames[i].astype(np.int16) - self.frames[i-1].astype(np.int16))
            diffs.append(np.mean(diff))
        elapsed = time.time() - start
        self.results['Frame differencing'] = f"{elapsed:.3f}s ({self.n_frames/elapsed:.1f} fps)"
        
        # Test 2: Color space conversion (RGB to grayscale)
        start = time.time()
        gray_frames = []
        for frame in self.frames:
            gray = np.dot(frame[...,:3], [0.299, 0.587, 0.114])
            gray_frames.append(gray.astype(np.uint8))
        elapsed = time.time() - start
        self.results['RGB to grayscale'] = f"{elapsed:.3f}s ({self.n_frames/elapsed:.1f} fps)"
        
        # Test 3: Frame scaling
        target_size = (640, 360)
        start = time.time()
        for frame in self.frames:
            img = Image.fromarray(frame)
            scaled = img.resize(target_size, Image.BILINEAR)
        elapsed = time.time() - start
        self.results['Frame scaling (50%)'] = f"{elapsed:.3f}s ({self.n_frames/elapsed:.1f} fps)"
        
        # Test 4: Gaussian blur (smoothing)
        start = time.time()
        for frame in self.frames[:50]:  # Sample subset for expensive operation
            from scipy.ndimage import gaussian_filter
            blurred = gaussian_filter(frame, sigma=2)
        elapsed = time.time() - start
        fps = 50 / elapsed
        self.results['Gaussian blur (sample)'] = f"{elapsed:.3f}s ({fps:.1f} fps)"
        
        # Test 5: Edge detection
        start = time.time()
        for gray_frame in gray_frames[:50]:
            edges = np.abs(np.diff(gray_frame, axis=0)) + np.abs(np.diff(gray_frame, axis=1))
        elapsed = time.time() - start
        fps = 50 / elapsed
        self.results['Edge detection (sample)'] = f"{elapsed:.3f}s ({fps:.1f} fps)"

if __name__ == '__main__':
    v = VideoProcessingBenchmark()
    v.setup()
    v.run()
