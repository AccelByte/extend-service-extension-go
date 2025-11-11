#!/bin/bash
set -e

echo "🚀 Setting up development environment..."

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod download

# Make scripts executable
echo "🔧 Setting up scripts..."
chmod +x proto.sh

# Generate protobuf files
echo "✏️ Generating protocol buffer files..."
if command -v protoc &> /dev/null; then
    if [ -d "pkg/proto" ]; then
        ./proto.sh || echo "⚠️  Protocol buffer generation skipped"
    else
        echo "⚠️  Proto directory not found, skipping generation"
    fi
else
    echo "⚠️  protoc not found"
fi

# Configure git for safe directory
if [ -d ".git" ]; then
    echo "🔧 Setting up git..."
    git config --global --add safe.directory /workspace
fi

echo "✅ Development environment setup complete!"
echo ""
echo "🎯 Quick start commands:"
echo "  • Run Go service: set -a && source .env && set +a && go run main.go"
echo "  • Build Go gateway: cd gateway && go build"
echo "  • Generate protobuf: ./proto.sh"
echo ""
echo "🛟 Ports:"
echo "  • gRPC Server: 6565"
echo "  • gRPC Gateway: 8000"
echo "  • Prometheus Metrics: 8080"
