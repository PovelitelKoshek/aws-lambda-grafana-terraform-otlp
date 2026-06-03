#!/bin/bash
set -e

echo "Cleaning old local build files..."
rm -f bootstrap lambda.zip

export GOCACHE="/tmp/gocache-lambda"
export GOMODCACHE="/tmp/gomodcache-lambda"
export GOTMPDIR="/tmp"

mkdir -p "$GOCACHE"
mkdir -p "$GOMODCACHE"

echo "Using GOCACHE=$GOCACHE"
echo "Using GOMODCACHE=$GOMODCACHE"

echo "Downloading Go dependencies..."
go mod tidy

echo "Building Lambda binary..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
go build -trimpath -ldflags="-s -w" -tags lambda.norpc -o bootstrap main.go

echo "Creating zip package..."
zip -j lambda.zip bootstrap

echo "Build complete:"
ls -lh bootstrap lambda.zip
