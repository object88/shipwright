#!/usr/bin/env bash
set -e

cd $(dirname "$0")

export TAG=${TAG:-$(git describe --tags)-$(git rev-parse --short HEAD)}

# Set defaults, allow env val to override
BUILD_AND_RELEASE=${BUILD_AND_RELEASE:-"false"}
DO_PUSH=${DO_PUSH:-"false"}
DO_TEST=${DO_TEST:-"true"}
DO_VERIFY=${DO_VERIFY:-"true"}
DO_VET=${DO_VET:-"true"}

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --fast)
      DO_TEST="false"
      DO_VERIFY="false"
      DO_VET="false"
      shift
      ;;
    --push)
      DO_PUSH="true"
      shift
      ;;
    --no-test)
      DO_TEST="false"
      shift
      ;;
    --no-verify)
      DO_VERIFY="false"
      shift
      ;;
    --no-vet)
      DO_VET="false"
      shift
      ;;
  esac
done

# Let this run the `go mod verify` task, so that we don't have to in every
# docker build.
time docker build \
  --build-arg DO_TEST=$DO_TEST \
  --build-arg DO_VET=$DO_VET \
  --build-arg DO_VERIFY=$DO_VERIFY \
  --build-arg BUILD_AND_RELEASE=$BUILD_AND_RELEASE \
  -f Dockerfile \
  --tag object88/shipwright:latest \
  .

docker tag "object88/shipwright:latest" "object88/shipwright:$TAG"

if [[ $DO_PUSH == "true" ]]; then
  echo "Pushing build images..."
  time docker push "object88/shipwright:$TAG"
  echo "Pushed images."
fi

echo "Finished building the docker images"
