#!/bin/bash

set -e

NGINX_URL="http://localhost"

echo "Testing infrastructure..."

# Wait for services
echo "Waiting for services to be healthy..."
sleep 10

# Test NGINX health
echo ""
echo "Testing NGINX health..."
if curl -f -s "${NGINX_URL}/health" > /dev/null; then
    echo "NGINX is healthy"
else
    echo "Error: NGINX health check failed"
    exit 1
fi

# Test event submission
echo ""
echo "Sending test event..."
RESPONSE=$(curl -s -X POST "${NGINX_URL}/events" \
    -H "Content-Type: application/json" \
    -d '{
        "device": "test-device",
        "os": "linux",
        "tipo": "test",
        "valor": "123",
        "ip": "127.0.0.1",
        "region": "us-east-1"
    }')

if echo "$RESPONSE" | grep -q "accepted"; then
    echo "Event accepted"
else
    echo "Error: Event submission failed"
    echo "Response: $RESPONSE"
    exit 1
fi

# Check container health
echo ""
echo "Checking container health..."
docker-compose ps

echo ""
echo "All tests passed!"
