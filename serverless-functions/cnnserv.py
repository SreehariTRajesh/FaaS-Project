from flask import Flask, request, jsonify
import torch
from torchvision import transforms, models
from PIL import Image
import time
import io
import base64

# Initialize Flask only once
app = Flask(__name__)

class CNNClassifier:
    def __init__(self, model_name="resnet50", device=None):
        # 1. Set device correctly (use CPU if not specified)
        self.device = device if device else "cpu"
        self.model = self._load_model(model_name)
        self.model.to(self.device) # Move model to device
        self.model.eval() # Set to evaluation mode! (Crucial for Batchnorm/Dropout)
        
        self.preprocess = self._get_preprocessing_pipeline()
        # You would typically load class names here
        self.class_names = {} 

    @staticmethod
    def _load_model(model_name):
        # Note: 'pretrained' is deprecated in newer versions, use 'weights' parameters if possible
        if model_name == "resnet50":
            model = models.resnet50(pretrained=True)
        elif model_name == "resnet18":
            model = models.resnet18(pretrained=True)
        elif model_name == "vgg16":
            model = models.vgg16(pretrained=True)
        else:
            raise ValueError(f"Unknown model: {model_name}")
        return model

    @staticmethod
    def _get_preprocessing_pipeline():
        return transforms.Compose([
            transforms.Scale(256), # Fixed: Scale is deprecated
            transforms.CenterCrop(224),
            transforms.ToTensor(),
            transforms.Normalize(
                mean=[0.485, 0.456, 0.406], 
                std=[0.229, 0.224, 0.225]
            ),
        ])

    def preprocess_image(self, image):
        image = image.convert("RGB")
        # Ensure input tensor is moved to the same device as the model
        return self.preprocess(image).unsqueeze(0).to(self.device)

    def predict(self, img_tensor, top_k=5):
        with torch.no_grad():
            outputs = self.model(img_tensor)
            probabilities = torch.nn.functional.softmax(outputs[0], dim=0)
            top_prob, top_idx = torch.topk(probabilities, top_k)
            predictions = []
            for i in range(top_k):
                predictions.append({
                    "class_id": top_idx[i].item(),
                    "class_name": self.class_names.get(top_idx[i].item(), f"Class {top_idx[i].item()}"),
                    "confidence": round(top_prob[i].item(), 4),
                })
            return predictions

# Initialize classifier globally
# NOTE: If deploying with Gunicorn, move this INSIDE the predict function
# or use a Lazy Loading pattern to avoid worker fork deadlocks.
classifier = CNNClassifier(model_name="resnet50")

@app.route("/health", methods=["GET"])
def health():
    return jsonify({"status": "healthy", "model": "resnet50", "device": str(classifier.device)}), 200

@app.route("/predict", methods=["POST"])
def predict():
    try:
        img = None # Initialize variable to avoid UnboundLocalError
        
        # 1. Handle File Upload
        if "image" in request.files:
            img_file = request.files["image"]
            img = Image.open(img_file.stream)
            
        # 2. Handle JSON Payload
        elif request.is_json and "image" in request.json:
            try:
                img_data = base64.b64decode(request.json["image"])
                img = Image.open(io.BytesIO(img_data))
            except Exception:
                return jsonify({"error": "Invalid base64 image data"}), 400

        # 3. Handle Missing Image
        if img is None:
             return jsonify({"error": "no image provided"}), 400

        # Processing
        top_k = request.args.get("top_k", default=5, type=int)
        
        tensor = classifier.preprocess_image(image=img)
        
        start_time = time.time_ns()
        predictions = classifier.predict(img_tensor=tensor, top_k=top_k)
        end_time = time.time_ns()
        
        # Convert nanoseconds to milliseconds for readability
        duration_ms = (end_time - start_time) / 1_000_000 
        
        return jsonify({
            "success": True, 
            "predictions": predictions, 
            "duration_ms": duration_ms
        }), 200

    except Exception as e:
        return jsonify({"success": False, "error": str(e)}), 500

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000, debug=False, threaded=False)