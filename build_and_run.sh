#!/bin/bash

# Extract port from config.yaml (specifically under server section)
PORT=$(grep -A 1 "server:" config.yaml | grep "port:" | awk -F'"' '{print $2}')
if [ -z "$PORT" ]; then
    PORT=8080
fi

echo "Killing process on port $PORT..."
fuser -k $PORT/tcp || true

# Update version and last_build in config.yaml
if [ -f "config.yaml" ]; then
  CURRENT_VERSION=$(grep "version:" config.yaml | head -n 1 | awk -F'"' '{print $2}')
  if [ ! -z "$CURRENT_VERSION" ]; then
    IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
    # Increment patch version
    NEW_VERSION="${VERSION_PARTS[0]}.${VERSION_PARTS[1]}.$((VERSION_PARTS[2] + 1))"
    CURRENT_DATE=$(date "+%Y-%m-%d %H:%M:%S")

    sed -i "s/version: \"$CURRENT_VERSION\"/version: \"$NEW_VERSION\"/" config.yaml
    sed -i "s/last_build: \".*\"/last_build: \"$CURRENT_DATE\"/" config.yaml

    echo "Updated config.yaml: version $NEW_VERSION, last_build $CURRENT_DATE"
  fi
fi

# Build the application
echo "Building GoBI..."
go build -o gobi cmd/gobi/main.go

if [ $? -eq 0 ]; then
    echo "Build successful. Starting GoBI on port $PORT..."
    
    # Load .env variables if file exists
    if [ -f .env ]; then
      # Export variables from .env, ignoring comments
      export $(grep -v '^#' .env | xargs)
    fi

    ./gobi
else
    echo "Build failed."
    exit 1
fi
