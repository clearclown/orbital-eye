"""
Orbital Eye — AI Worker (gRPC Server)
Handles ML inference for object detection, classification, change detection.
"""
import argparse
import io
import logging
import time
from concurrent import futures
from pathlib import Path

import cv2
import grpc
import numpy as np
from PIL import Image

import detector_pb2
import detector_pb2_grpc

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
logger = logging.getLogger("ai-worker")


class DetectorServicer(detector_pb2_grpc.DetectorServiceServicer):
    def __init__(self, model_dir: str = "models/", device: str = "auto"):
        self.model_dir = Path(model_dir)
        self.model_dir.mkdir(parents=True, exist_ok=True)
        self.models = {}

        import torch
        if device == "auto":
            self.device = "cuda" if torch.cuda.is_available() else "cpu"
        else:
            self.device = device
        logger.info(f"AI Worker: device={self.device}")

    def _get_detector(self, target_class: str = "general"):
        """Get or load a YOLO detection model."""
        if target_class in self.models:
            return self.models[target_class]

        from ultralytics import YOLO

        # Use pretrained YOLOv8 models
        # For satellite imagery, we start with general YOLO and will fine-tune later
        model_map = {
            "general": "yolov8m.pt",      # Medium model, good balance
            "vessels": "yolov8m.pt",       # TODO: fine-tuned vessel model
            "aircraft": "yolov8m.pt",      # TODO: fine-tuned aircraft model
        }

        model_name = model_map.get(target_class, "yolov8m.pt")
        custom_path = self.model_dir / model_name

        if custom_path.exists():
            logger.info(f"Loading custom model: {custom_path}")
            model = YOLO(str(custom_path))
        else:
            logger.info(f"Loading pretrained model: {model_name}")
            model = YOLO(model_name)

        model.to(self.device)
        self.models[target_class] = model
        return model

    def _load_image(self, request):
        """Load image from request (bytes or path)."""
        if request.image_path:
            img = cv2.imread(request.image_path)
            if img is None:
                raise ValueError(f"Cannot read image: {request.image_path}")
            return img
        elif request.image_data:
            nparr = np.frombuffer(request.image_data, np.uint8)
            img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
            if img is None:
                raise ValueError("Cannot decode image data")
            return img
        else:
            raise ValueError("No image provided")

    def DetectObjects(self, request, context):
        logger.info(f"DetectObjects: targets={list(request.target_classes)}, "
                     f"confidence={request.confidence_threshold}, gsd={request.gsd_meters}m")
        start = time.time()

        try:
            img = self._load_image(request)
        except ValueError as e:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details(str(e))
            return detector_pb2.DetectResponse()

        # Determine which model to use
        target = "general"
        if request.target_classes:
            target = request.target_classes[0]

        model = self._get_detector(target)
        conf = request.confidence_threshold if request.confidence_threshold > 0 else 0.25

        # Run inference
        results = model(img, conf=conf, verbose=False)

        detections = []
        for r in results:
            for box in r.boxes:
                cls_id = int(box.cls[0])
                cls_name = model.names[cls_id]
                conf_val = float(box.conf[0])
                x1, y1, x2, y2 = box.xyxy[0].tolist()

                det = detector_pb2.Detection(
                    class_name=cls_name,
                    confidence=conf_val,
                    bbox=detector_pb2.BoundingBox(
                        x_min=x1, y_min=y1, x_max=x2, y_max=y2
                    ),
                )

                # Estimate physical size if GSD is provided
                if request.gsd_meters > 0:
                    det.estimated_length_m = (x2 - x1) * request.gsd_meters
                    det.estimated_width_m = (y2 - y1) * request.gsd_meters

                # Geo-reference if top_left is provided
                if request.top_left.latitude != 0 or request.top_left.longitude != 0:
                    cx = (x1 + x2) / 2
                    cy = (y1 + y2) / 2
                    # Simple pixel → geo conversion
                    gsd_deg_lat = request.gsd_meters / 111320.0
                    gsd_deg_lon = request.gsd_meters / (111320.0 * np.cos(np.radians(request.top_left.latitude)))
                    det.geo_center.CopyFrom(detector_pb2.GeoPoint(
                        latitude=request.top_left.latitude - cy * gsd_deg_lat,
                        longitude=request.top_left.longitude + cx * gsd_deg_lon,
                    ))

                detections.append(det)

        elapsed = (time.time() - start) * 1000
        logger.info(f"  Found {len(detections)} objects in {elapsed:.0f}ms")

        return detector_pb2.DetectResponse(
            detections=detections,
            inference_time_ms=elapsed,
            model_version=f"yolov8m-{target}",
        )

    def DetectChanges(self, request, context):
        logger.info(f"DetectChanges: sensitivity={request.sensitivity}")
        start = time.time()

        try:
            if request.image_before_path and request.image_after_path:
                img1 = cv2.imread(request.image_before_path)
                img2 = cv2.imread(request.image_after_path)
            elif request.image_before and request.image_after:
                img1 = cv2.imdecode(np.frombuffer(request.image_before, np.uint8), cv2.IMREAD_COLOR)
                img2 = cv2.imdecode(np.frombuffer(request.image_after, np.uint8), cv2.IMREAD_COLOR)
            else:
                raise ValueError("Both before and after images required")

            if img1 is None or img2 is None:
                raise ValueError("Cannot load one or both images")
        except ValueError as e:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details(str(e))
            return detector_pb2.ChangeResponse()

        # Simple pixel-level change detection
        # TODO: Replace with trained Siamese UNet for better results
        if img1.shape != img2.shape:
            img2 = cv2.resize(img2, (img1.shape[1], img1.shape[0]))

        diff = cv2.absdiff(img1, img2)
        gray_diff = cv2.cvtColor(diff, cv2.COLOR_BGR2GRAY)

        sensitivity = request.sensitivity if request.sensitivity > 0 else 0.5
        threshold = int(255 * (1.0 - sensitivity))
        _, binary = cv2.threshold(gray_diff, threshold, 255, cv2.THRESH_BINARY)

        # Morphological cleanup
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (5, 5))
        binary = cv2.morphologyEx(binary, cv2.MORPH_OPEN, kernel)
        binary = cv2.morphologyEx(binary, cv2.MORPH_CLOSE, kernel)

        # Find contours (change regions)
        contours, _ = cv2.findContours(binary, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)

        regions = []
        total_pixels = img1.shape[0] * img1.shape[1]
        change_pixels = cv2.countNonZero(binary)

        for cnt in contours:
            area = cv2.contourArea(cnt)
            if area < 100:  # Skip tiny changes
                continue
            x, y, w, h = cv2.boundingRect(cnt)
            significance = min(1.0, area / 10000.0)
            regions.append(detector_pb2.ChangeRegion(
                bbox=detector_pb2.BoundingBox(x_min=x, y_min=y, x_max=x+w, y_max=y+h),
                change_type="activity_change",
                significance=significance,
            ))

        # Encode change mask
        _, mask_bytes = cv2.imencode('.png', binary)

        elapsed = (time.time() - start) * 1000
        logger.info(f"  Found {len(regions)} change regions ({change_pixels/total_pixels*100:.1f}% changed) in {elapsed:.0f}ms")

        return detector_pb2.ChangeResponse(
            change_mask=mask_bytes.tobytes(),
            regions=regions,
            change_percentage=change_pixels / total_pixels * 100,
        )

    def Health(self, request, context):
        import torch
        gpu_used, gpu_total = 0, 0
        if torch.cuda.is_available():
            gpu_used = torch.cuda.memory_allocated() // (1024 * 1024)
            props = torch.cuda.get_device_properties(0)
            gpu_total = props.total_mem // (1024 * 1024)

        return detector_pb2.HealthResponse(
            ready=True,
            loaded_models=list(self.models.keys()),
            gpu_memory_used_mb=gpu_used,
            gpu_memory_total_mb=gpu_total,
        )

    def Enhance(self, request, context):
        # TODO: Implement super-resolution
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details("Super-resolution not yet implemented")
        return detector_pb2.EnhanceResponse()


def serve(port: int = 50051, model_dir: str = "models/", device: str = "auto"):
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=4),
        options=[
            ('grpc.max_receive_message_length', 100 * 1024 * 1024),  # 100MB
            ('grpc.max_send_message_length', 100 * 1024 * 1024),
        ]
    )
    servicer = DetectorServicer(model_dir=model_dir, device=device)
    detector_pb2_grpc.add_DetectorServiceServicer_to_server(servicer, server)
    server.add_insecure_port(f"[::]:{port}")
    server.start()
    logger.info(f"AI Worker listening on port {port}")
    server.wait_for_termination()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Orbital Eye AI Worker")
    parser.add_argument("--port", type=int, default=50051)
    parser.add_argument("--model-dir", default="models/")
    parser.add_argument("--device", default="auto", choices=["auto", "cuda", "cpu"])
    args = parser.parse_args()
    serve(port=args.port, model_dir=args.model_dir, device=args.device)
