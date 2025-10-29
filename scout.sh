#!/bin/bash

set -e

echo "Running Docker Scout analysis..."

# Check if Docker Scout is available
if ! docker scout version &> /dev/null; then
    echo "Error: Docker Scout is not available"
    echo "Install with: docker scout install"
    exit 1
fi

echo ""
echo "Analyzing Producer image..."
docker scout cves producer:latest

echo ""
echo "Analyzing Consumer image..."
docker scout cves consumer:latest

echo ""
echo "Getting recommendations for Producer..."
docker scout recommendations producer:latest

echo ""
echo "Getting recommendations for Consumer..."
docker scout recommendations consumer:latest

echo ""
echo "Scout analysis completed!"
