#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

printf "Starting connection test server %s.\n" "$HOST_NAME" > _connTestServer.log

./_connTest -name "$HOST_NAME" -port 27271 > _connTestServer.log 2> _connTestServer.err
