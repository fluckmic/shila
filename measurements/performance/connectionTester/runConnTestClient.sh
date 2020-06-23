#!/bin/bash

HOST_NAME=$(uname -n)
HOST_ID=$(cat _HOST_ID)
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

CLIENT_ADDR=(17-ffaa:1:d87 19-ffaa:1:d88 20-ffaa:1:d89 18-ffaa:1:d8a)

printf "Starting connection test client %s with id %d.\n " "$HOST_NAME" "$HOST_ID"

#.././_connTest -name "$HOST_NAME" -port 27271 &

exit $?

