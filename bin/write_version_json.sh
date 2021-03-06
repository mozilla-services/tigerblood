#!/usr/bin/env bash
set -eo pipefail

## writes a version.json file for the /__version__ endpoint
# per https://github.com/mozilla-services/Dockerflow/blob/master/docs/version_object.md

# default variables
: "${CIRCLE_SHA1=$(git rev-parse HEAD)}"
: "${CIRCLE_TAG=$(git describe --tags)}"
: "${CIRCLE_PROJECT_USERNAME=mozilla-services}"
: "${CIRCLE_PROJECT_REPONAME=tigerblood}"
: "${CIRCLE_BUILD_URL=localdev}"

rm -f version.json

printf '{"commit":"%s","version":"%s","source":"https://github.com/%s/%s","build":"%s"}\n' \
            "$CIRCLE_SHA1" \
            "$CIRCLE_TAG" \
            "$CIRCLE_PROJECT_USERNAME" \
            "$CIRCLE_PROJECT_REPONAME" \
            "$CIRCLE_BUILD_URL" > version.json
