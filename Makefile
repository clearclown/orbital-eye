.PHONY: build proto ai-deps ai-serve test clean

# Build Go binary
build:
	go build -o bin/orbital-eye ./cmd/orbital-eye/

# Generate gRPC code from proto
proto:
	protoc --go_out=. --go-grpc_out=. proto/detector.proto
	cd ai && python -m grpc_tools.protoc -I../proto --python_out=. --grpc_python_out=. ../proto/detector.proto

# Install Python AI dependencies
ai-deps:
	cd ai && pip install -r requirements.txt

# Start AI worker
ai-serve:
	cd ai && python server.py --port 50051

# Download training datasets
datasets:
	./scripts/download_datasets.sh data/training

# Run tests
test:
	go test ./...
	cd ai && python -m pytest tests/

# Clean
clean:
	rm -rf bin/ proto/gen/
