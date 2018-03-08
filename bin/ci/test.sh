#!/usr/bin/env bash
set -eo pipefail

CI_ENV=`bash <(curl -s https://codecov.io/env)`
echo 'CI_ENV:' $CI_ENV
docker-compose run $CI_ENV test test-ci
