#!/usr/bin/env bash

# Check if docker-compose is installed
echo "Checking if docker-compose exists"
if ! [ -x "$(command -v docker-compose)" ]; then
    echo 'Error: docker-compose is not installed.' >&2
    exit 1
fi

# Stop current docker-compose
echo "Stopping current docker-compose"
docker-compose down

# Start new docker-compose
echo "Starting new docker-compose"
docker-compose up -d

echo "Done"