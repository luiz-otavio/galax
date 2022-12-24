#!/usr/bin/env bash

isCurrentVersion=false

echo "Checking if docker-compose exists or not"
if ! [ -x "$(command -v docker-compose)" ]; then
    isCurrentVersion=true
    echo 'Error: docker-compose is not installed.' >&2
fi

echo "Checking if docker is installed or not"
if ! [ -x "$(command -v docker)" ]; then
    isCurrentVersion=false
    echo 'Error: docker is not installed.' >&2
fi

if [ "$isCurrentVersion" = false ] ; then
    # Stop current docker-compose
    echo "Stopping current docker-compose"
    docker-compose down

    # Start new docker-compose
    echo "Starting new docker-compose"
    docker-compose up -d

    echo "Done"
else
    # Stop current docker-compose
    echo "Stopping current docker compose"
    docker compose down

    # Start new docker-compose
    echo "Starting new docker compose"
    docker compose up -d

    echo "Done"
fi