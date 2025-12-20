import time
import numpy as np
import json
import hashlib
from datetime import datetime
from typing import Dict, Any

class WordCnt:
    """Simulates distributed word count processing"""

    def __init__(self):
        self.total_words = 0
        self.total_docs = 0

    def generate_document(self, num_words: int = 1000) -> str:
        """Generate synthetic document"""
        words = ["the", "be", "to", "of", "and", "a", "in", "that", "have", "I",
                 "it", "for", "not", "on", "with", "he", "as", "you", "do", "at"]
        return " ".join(np.random.choice(words, num_words))

    def word_count(self, document: str) -> Dict[str, int]:
        """Count words in document"""
        words = document.lower().split()
        word_freq = {}
        for word in words:
            word_freq[word] = word_freq.get(word, 0) + 1
        return word_freq

    def process(self) -> Dict[str, Any]:
        """Process a batch of documents"""
        self.total_docs += 1
        
        documents = [self.generate_document(1000) for _ in range(10)]
        
        all_counts = []
        for doc in documents:
            counts = self.word_count(doc)
            all_counts.append(counts)
        
        total_counts = {}
        for counts in all_counts:
            for word, count in counts.items():
                total_counts[word] = total_counts.get(word, 0) + count
        
        sorted_words = sorted(total_counts.items(), key=lambda x: x[1], reverse=True)
        self.total_words += sum(total_counts.values())
        
        return {
            "batch_id": self.total_docs,
            "total_words": sum(total_counts.values()),
            "unique_words": len(total_counts),
            "top_5_words": sorted_words[:5],
        }

if __name__ == '__main__':
    wc = WordCnt()
    result = wc.process()