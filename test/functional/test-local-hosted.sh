#!/usr/bin/env bash

# Prerequisites: bash, curl, go, jq

set -e
set -o pipefail
#set -x

function clean_up()
{
  kill -9 $SERVICE_PID
}

trap clean_up EXIT

echo '# Build and run Extend app locally'

go build -o service
./service & SERVICE_PID=$!

(for _ in {1..12}; do bash -c "timeout 1 echo > /dev/tcp/127.0.0.1/8080" 2>/dev/null && exit 0 || sleep 5s; done; exit 1)

if [ $? -ne 0 ]; then
  echo "Failed to run Extend app locally"
  exit 1
fi

echo '# Testing Extend app using demo script'

bash demo.sh
