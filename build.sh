#!/bin/sh
set -e

BIN_PATH=$(dirname "$0")
REPO=$(cd "$BIN_PATH"; pwd)
COMMIT_SHA=$(git rev-parse --short HEAD)
BUILD_DATE=$(date +%s)
VERSION=1.0.0

cd "$REPO"

go build -a -o journey -ldflags " \
    -X 'journey/cardinal.Version=$VERSION' \
    -X 'journey/cardinal.LastCommit=$COMMIT_SHA' \
    -X 'journey/cardinal.BuildDate=$BUILD_DATE' \
    "
