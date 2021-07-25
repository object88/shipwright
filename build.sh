#!/usr/bin/env bash

set -e

DEFAULT_GOOS=$(uname | tr '[:upper:]' '[:lower:]')
PLATFORMS=( "$DEFAULT_GOOS/amd64" )
if [ "$BUILD_AND_RELEASE" == "true" ]; then
  PLATFORMS=( "linux/amd64" "darwin/amd64" )
fi

for PLATFORM in "${PLATFORMS[@]}"; do
  export GOOS=$(cut -d'/' -f1 <<< $PLATFORM)
  export GOARCH=$(cut -d'/' -f2 <<< $PLATFORM)
  BINARY_NAME="shipwright-${GOOS}-${GOARCH}"

  go build -o ./bin/${BINARY_NAME} ./main/main.go
done