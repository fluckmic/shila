#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

# Load the addresses of the scion access points.
ADDRESSES=(17-ffaa:1:d87,[127.0.0.1]:27271 19-ffaa:1:d88,[127.0.0.1]:27271)
OWN_ADDR="${ADDRESSES["$HOST_ID" - 1]}"

printf "Starting connection test client %s.\n" "$HOST_NAME" >> _connTestClient.log

for ADDRESS in "${ADDRESSES[@]}"; do
  if [[ "$ADDRESS" != "$OWN_ADDR" ]]; then
    ./_connTest -name "$HOST_NAME" -remote "$ADDRESS" >> _connTestClient.log 2> _connTestClient.error
  fi
done

exit 0

