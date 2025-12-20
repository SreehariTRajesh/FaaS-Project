import time 
from PIL import Image
import numpy as np 
import io

class ThumbnailBenchmark():
    """Benchmark for image thumbnail generation"""
    
    def __init__(self, n_images=100, size=(1920, 1080), thumb_size=(200, 150)):
        self.n_images = n_images
        self.size = size
        self.thumb_size = thumb_size
        self.images = []
        self.results = {}
        
    def setup(self):
        """Generate synthetic test images"""
        print(f"Generating {self.n_images} test images ({self.size[0]}x{self.size[1]})...")
        for i in range(self.n_images):
            arr = np.random.randint(0, 255, (self.size[1], self.size[0], 3), dtype=np.uint8)
            self.images.append(Image.fromarray(arr))
    
    def run(self):
        """Run thumbnail generation benchmarks"""
        # Test 1: Basic resize
        start = time.time()
        thumbnails = []
        for img in self.images:
            thumb = img.resize(self.thumb_size, Image.LANCZOS)
            thumbnails.append(thumb)
        elapsed = time.time() - start
        self.results['Basic resize'] = f"{elapsed:.3f}s ({self.n_images/elapsed:.1f} imgs/sec)"
        
        # Test 2: Resize + enhancement
        start = time.time()
        for img in self.images:
            thumb = img.resize(self.thumb_size, Image.LANCZOS)
            thumb = thumb.convert('RGB')
            enhancer = np.array(thumb)
            enhancer = np.clip(enhancer * 1.2, 0, 255).astype(np.uint8)
            thumb = Image.fromarray(enhancer)
        elapsed = time.time() - start
        self.results['Resize + enhancement'] = f"{elapsed:.3f}s ({self.n_images/elapsed:.1f} imgs/sec)"
        
        # Test 3: Resize + save to buffer
        start = time.time()
        for img in self.images:
            thumb = img.resize(self.thumb_size, Image.LANCZOS)
            buffer = io.BytesIO()
            thumb.save(buffer, format='JPEG', quality=85)
        elapsed = time.time() - start
        self.results['Resize + JPEG encode'] = f"{elapsed:.3f}s ({self.n_images/elapsed:.1f} imgs/sec)"
        
        # Test 4: Batch processing with cropping
        start = time.time()
        for img in self.images:
            # Center crop to square
            width, height = img.size
            min_dim = min(width, height)
            left = (width - min_dim) // 2
            top = (height - min_dim) // 2
            img_cropped = img.crop((left, top, left + min_dim, top + min_dim))
            thumb = img_cropped.resize(self.thumb_size, Image.LANCZOS)
        elapsed = time.time() - start
        self.results['Crop + resize'] = f"{elapsed:.3f}s ({self.n_images/elapsed:.1f} imgs/sec)"


if __name__ == '__main__':
    t = ThumbnailBenchmark()
    t.setup()
    t.run()
