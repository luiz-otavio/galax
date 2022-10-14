#!/usr/bin/env bash

# Check if proto-gen-go is installed
echo "Checking if goreleaser exists"
if ! [ -x "$(command -v goreleaser)" ]; then
    # Check if go is installed
    if ! [ -x "$(command -v go)" ]; then
        echo 'Error: go is not installed.' >&2
        exit 1
    fi

    # Install goreleaser
    go install github.com/goreleaser/goreleaser@latest

    echo "goreleaser is installed."
fi

goreleaser build --rm-dist --snapshot -p 2

echo "Done"
