#!/usr/bin/env bash
set -eo pipefail

# default variables
: "${PORT:=8080}"
: "${SLEEP:=1}"
: "${TRIES:=60}"

usage() {
  echo "usage: ./bin/run.sh web|test|test-ci|bash"
  exit 1
}

wait_for() {
  tries=0
  echo "Waiting for $1 to listen on $2..."
  while true; do
    [[ $tries -lt $TRIES ]] || return
    (echo > /dev/tcp/$1/$2) >/dev/null 2>&1
    result=
    [[ $? -eq 0 ]] && return
    sleep $SLEEP
    tries=$((tries + 1))
  done
}

[ $# -lt 1 ] && usage

# Only wait for backend services in development
# http://stackoverflow.com/a/13864829
# For example, bin/test.sh sets 'DEVELOPMENT' to something
[ ! -z ${DEVELOPMENT+check} ] && wait_for db 5432 && sleep 3

case $1 in
  web)
    /go/bin/tigerblood "${@:2}"
    ;;
  test)
    cd /go/src/go.mozilla.org/tigerblood
    go test -v ./
    ;;
  test-ci)
    cd /go/src/go.mozilla.org/tigerblood
    go test -v ./ -coverprofile=/tmp/coverage.txt -covermode=atomic
    bash <(curl -s https://codecov.io/bash) -s /tmp/coverage.txt
    ;;
  bash)
    exec "$@"
    ;;
  *)
    ;;
esac
