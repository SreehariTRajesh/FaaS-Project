import time
import numpy as np
import json
import hashlib
import base64
from datetime import datetime
from typing import Dict, Any
import sys


class LrSrv:
    """Simulates logistic regression training and inference"""

    def __init__(self, n_features: int = 100, n_samples: int = 1000):
        self.n_features = n_features
        self.n_samples = n_samples
        self.weights = np.random.randn(n_features)
        self.bias = 0.0
        self.training_count = 0

    def sigmoid(self, z: np.ndarray) -> np.ndarray:
        """Sigmoid activation"""
        return 1 / (1 + np.exp(-np.clip(z, -500, 500)))

    def train_epoch(self, X: np.ndarray, y: np.ndarray, learning_rate: float = 0.01):
        """Train for one epoch"""
        # Forward pass
        z = np.dot(X, self.weights) + self.bias
        predictions = self.sigmoid(z)

        # Compute gradients
        dw = (1 / self.n_samples) * np.dot(X.T, (predictions - y))
        db = (1 / self.n_samples) * np.sum(predictions - y)

        # Update weights
        self.weights -= learning_rate * dw
        self.bias -= learning_rate * db

        # Compute loss
        loss = -(1 / self.n_samples) * np.sum(
            y * np.log(predictions + 1e-8) + (1 - y) * np.log(1 - predictions + 1e-8)
        )

        return loss

    def train_model(self, epochs: int = 10) -> Dict[str, Any]:
        """Train logistic regression model"""
        self.training_count += 1

        # Generate synthetic data
        X = np.random.randn(self.n_samples, self.n_features)
        true_weights = np.random.randn(self.n_features)
        y = (
            np.dot(X, true_weights) + np.random.randn(self.n_samples) * 0.1 > 0
        ).astype(float)

        # Train for multiple epochs
        losses = []
        for epoch in range(epochs):
            loss = self.train_epoch(X, y)
            losses.append(loss)

        # Compute accuracy on training data
        predictions = self.sigmoid(np.dot(X, self.weights) + self.bias)
        accuracy = np.mean((predictions > 0.5) == y)

        return {
            "training_id": self.training_count,
            "epochs": epochs,
            "final_loss": losses[-1],
            "accuracy": float(accuracy),
            "weight_norm": float(np.linalg.norm(self.weights)),
        }

    def run_continuous(self, duration: float = 60.0):
        """Run continuously for specified duration"""
        print(f"[LrSrv] Starting continuous execution for {duration}s")
        start_time = time.time()
        iterations = 0

        while time.time() - start_time < duration:
            result = self.train_model(epochs=10)
            iterations += 1

            if iterations % 5 == 0:
                elapsed = time.time() - start_time
                print(
                    f"[LrSrv] Completed {iterations} training runs in {elapsed:.2f}s "
                    f"({iterations/elapsed:.2f} runs/s)"
                )

        total_time = time.time() - start_time
        print(f"[LrSrv] Completed {iterations} iterations in {total_time:.2f}s")
        return iterations


if __name__ == "__main__":
    lrsrv = LrSrv(n_features=20, n_samples=1000)
    lrsrv.run_continuous(duration=60)
