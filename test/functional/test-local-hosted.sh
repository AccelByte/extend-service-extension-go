#!/usr/bin/env bash

# Prerequisites: bash, curl, go, jq

set -e
set -o pipefail
#set -x

APP_BASE_URL=http://localhost:8000
APP_BASE_PATH="guild"

function clean_up()
{
  kill -9 $SERVICE_PID
}

trap clean_up EXIT

echo '# Build and run Extend app locally'

sed -i "s@base_path:[ ]*\"[^\"]\+\"@base_path: \"/${APP_BASE_PATH}\"@" \
    pkg/proto/guildService.proto
sed -i "s@BasePath[ ]*=[ ]*\"[^\"]\+\"@BasePath = \"/${APP_BASE_PATH}\"@" \
    pkg/common/config.go

go build -o service
BASE_PATH=/$APP_BASE_PATH ./service & SERVICE_PID=$!

(for _ in {1..12}; do bash -c "timeout 1 echo > /dev/tcp/127.0.0.1/8080" 2>/dev/null && exit 0 || sleep 5s; done; exit 1)

if [ $? -ne 0 ]; then
  echo "Failed to run Extend app locally"
  exit 1
fi

echo '# Testing Extend app using demo script'

export SERVICE_BASE_URL=$APP_BASE_URL
export SERVICE_BASE_PATH=$APP_BASE_PATH

bash demo.sh
