#!/bin/bash

# THIS IS MEANT TO BE RUN BY CI

set -e

# Usage: retry MAX CMD...
# Retry CMD up to MAX times. If it fails MAX times, returns failure.
# Example: retry 3 docker push "mozilla/tigerblood:$TAG"
function retry() {
    max=$1
    shift
    count=1
    until "$@"; do
        count=$((count + 1))
        if [[ $count -gt $max ]]; then
            return 1
        fi
        echo "$count / $max"
    done
    return 0
}

# configure docker creds
retry 3  echo "$DOCKER_PASS" | docker login -u="$DOCKER_USER" --password-stdin

# docker tag and push git branch to dockerhub
if [ -n "$1" ]; then
    [ "$1" == master ] && TAG=latest || TAG="$1"
    docker tag tigerblood:build "mozilla/tigerblood:$TAG" ||
        (echo "Couldn't tag tigerblood:build as mozilla/tigerblood:$TAG" && false)
    retry 3 docker push "mozilla/tigerblood:$TAG" ||
        (echo "Couldn't push mozilla/tigerblood:$TAG" && false)
    echo "Pushed mozilla/tigerblood:$TAG"

    docker tag tigerblood:test_db "mozilla/tigerblood_test_db:$TAG" ||
        (echo "Couldn't tag tigerblood:test_db as mozilla/tigerblood_test_db:$TAG" && false)
    retry 3 docker push "mozilla/tigerblood_test_db:$TAG" ||
        (echo "Couldn't push mozilla/tigerblood_test_db:$TAG" && false)
fi
