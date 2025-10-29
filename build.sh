#!/bin/bash

set -e

echo "Building Docker images with Buildx..."

# Check if buildx is available
if ! docker buildx version &> /dev/null; then
    echo "Error: Docker Buildx is not available"
    exit 1
fi

# Create builder if not exists
if ! docker buildx inspect multiarch &> /dev/null; then
    echo "Creating buildx builder..."
    docker buildx create --name multiarch --use
fi

# Build with bake
echo "Building with docker bake..."
docker buildx bake --load

echo "Build completed!"

# Generate SBOM
echo "Generating SBOM..."
mkdir -p sbom
docker buildx bake producer-sbom consumer-sbom

echo "SBOM generated in ./sbom/"

# List images
echo ""
echo "Built images:"
docker images | grep -E "producer|consumer"
