#!/bin/bash

HOST_NAME=$(uname -n)

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

HOST_ID=$(cat ../_HOST_ID)

CLIENT_ADDR=(17-ffaa:1:d87,[127.0.0.1]:27271 19-ffaa:1:d88,[127.0.0.1]:27271 20-ffaa:1:d89,[127.0.0.1]:27271 18-ffaa:1:d8a,[127.0.0.1]:27271)

printf "Starting connection test client %s.\n" "$HOST_NAME"
for i in {1..4}; do
  if [[ "$i" -ne "$HOST_ID" ]]; then
  .././_connTest -name "$HOST_NAME" -remote "${CLIENT_ADDR["$i"-1]}"
    if [[  $? -ne 0 ]]; then
    continue
    fi
  fi
done

exit 0

