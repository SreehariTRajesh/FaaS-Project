import numpy as np 
import time
from collections import deque, defaultdict

class GraphProcessingBenchmark():
    """Benchmark for graph algorithms (BFS, PageRank)"""
    
    def __init__(self, n_nodes=10000, edge_probability=0.001):
        self.n_nodes = n_nodes
        self.edge_probability = edge_probability
        self.graph = None
        
    def setup(self):
        """Generate random graph"""
        print(f"Generating random graph with {self.n_nodes} nodes...")
        self.graph = defaultdict(list)
        edge_count = 0
        
        # Generate random edges
        for i in range(self.n_nodes):
            for j in range(i + 1, self.n_nodes):
                if np.random.random() < self.edge_probability:
                    self.graph[i].append(j)
                    self.graph[j].append(i)
                    edge_count += 1
        
        print(f"Generated {edge_count} edges (avg degree: {2*edge_count/self.n_nodes:.1f})")
    
    def bfs(self, start_node):
        """Breadth-first search"""
        visited = set()
        queue = deque([start_node])
        visited.add(start_node)
        
        while queue:
            node = queue.popleft()
            for neighbor in self.graph[node]:
                if neighbor not in visited:
                    visited.add(neighbor)
                    queue.append(neighbor)
        
        return visited
    
    def pagerank(self, iterations=20, damping=0.85):
        """Simple PageRank implementation"""
        n = self.n_nodes
        ranks = {i: 1.0 / n for i in range(n)}
        
        for _ in range(iterations):
            new_ranks = {}
            for node in range(n):
                rank_sum = 0.0
                # Find all nodes that link to this node
                for source in range(n):
                    if node in self.graph[source]:
                        out_degree = len(self.graph[source])
                        if out_degree > 0:
                            rank_sum += ranks[source] / out_degree
                
                new_ranks[node] = (1 - damping) / n + damping * rank_sum
            
            ranks = new_ranks
        
        return ranks
    
    def run(self):
        """Run graph processing benchmarks"""
        # Test 1: BFS from single node
        start = time.time()
        visited = self.bfs(0)
        elapsed = time.time() - start
        self.results['BFS (single source)'] = f"{elapsed:.3f}s, visited {len(visited)} nodes"
        
        # Test 2: BFS from multiple nodes
        start = time.time()
        for start_node in range(0, min(100, self.n_nodes), 10):
            visited = self.bfs(start_node)
        elapsed = time.time() - start
        self.results['BFS (10 sources)'] = f"{elapsed:.3f}s"
        
        # Test 3: PageRank (10 iterations)
        start = time.time()
        ranks = self.pagerank(iterations=10)
        elapsed = time.time() - start
        top_nodes = sorted(ranks.items(), key=lambda x: x[1], reverse=True)[:5]
        self.results['PageRank (10 iter)'] = f"{elapsed:.3f}s, top node: {top_nodes[0][0]} ({top_nodes[0][1]:.6f})"
        
        # Test 4: PageRank (20 iterations)
        start = time.time()
        ranks = self.pagerank(iterations=20)
        elapsed = time.time() - start
        self.results['PageRank (20 iter)'] = f"{elapsed:.3f}s"
        
        # Test 5: Connected components (using BFS)
        start = time.time()
        visited_global = set()
        components = 0
        for node in range(self.n_nodes):
            if node not in visited_global:
                component = self.bfs(node)
                visited_global.update(component)
                components += 1
        elapsed = time.time() - start
        self.results['Connected components'] = f"{elapsed:.3f}s, found {components} components"

if __name__ == '__main__':
    graphproc = GraphProcessingBenchmark()
    graphproc.setup()
    graphproc.run()