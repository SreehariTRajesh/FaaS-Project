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
# 5. WordCnt - Word Count (Map-Reduce Style)
# ============================================================================
class WordCnt:
    """Simulates distributed word count processing"""

    def __init__(self):
        self.total_words = 0
        self.total_docs = 0

    def generate_document(self, num_words: int = 1000) -> str:
        """Generate synthetic document"""
        words = [
            "the",
            "be",
            "to",
            "of",
            "and",
            "a",
            "in",
            "that",
            "have",
            "I",
            "it",
            "for",
            "not",
            "on",
            "with",
            "he",
            "as",
            "you",
            "do",
            "at",
            "this",
            "but",
            "his",
            "by",
            "from",
            "they",
            "we",
            "say",
            "her",
            "she",
        ]

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

        # Generate multiple documents
        documents = [self.generate_document(1000) for _ in range(10)]

        # Map phase: count words in each document
        all_counts = []
        for doc in documents:
            counts = self.word_count(doc)
            all_counts.append(counts)

        # Reduce phase: aggregate counts
        total_counts = {}
        for counts in all_counts:
            for word, count in counts.items():
                total_counts[word] = total_counts.get(word, 0) + count

        # Sort by frequency
        sorted_words = sorted(total_counts.items(), key=lambda x: x[1], reverse=True)

        self.total_words += sum(total_counts.values())

        return {
            "batch_id": self.total_docs,
            "total_words": sum(total_counts.values()),
            "unique_words": len(total_counts),
            "top_5_words": sorted_words[:5],
        }

    def run_continuous(self, duration: float = 60.0):
        """Run continuously for specified duration"""
        print(f"[WordCnt] Starting continuous execution for {duration}s")
        start_time = time.time()
        iterations = 0

        while time.time() - start_time < duration:
            result = self.process()
            iterations += 1

            if iterations % 10 == 0:
                elapsed = time.time() - start_time
                print(
                    f"[WordCnt] Processed {iterations} batches in {elapsed:.2f}s "
                    f"({iterations/elapsed:.2f} batch/s)"
                )

        total_time = time.time() - start_time
        print(f"[WordCnt] Completed {iterations} iterations in {total_time:.2f}s")
        return iterations


if __name__ == "__main__":
    wcnt = WordCnt()
    wcnt.run_continuous(duration=60)
