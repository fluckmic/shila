#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

# Load the addresses of the scion access points.
ADDRESSES="${cat scionApAddresses}"
OWN_ADDR=$(ADDRESSES["$HOST_ID"])


printf "Starting connection test client %s.\n" "$HOST_NAME" > _connTestClient.log
printf "Own addr: %s.\n" "$OWN_ADDR" > _connTestClient.log

for ADDRESS in $(ADDRESSES[@]); do
  printf "Connection test to %s.\n" "$ADDRESS" > _connTestClient.log
  if [[ "$ADDRESS" != "$OWN_ADDR" ]]; then
    printf "Connection test to %s.\n" "$ADDRESS" > _connTestClient.log
    ./_connTest -name "$HOST_NAME" -remote "$ADDRESS" > _connTestClient.log 2> _connTestClient.error
  fi
done

exit 0

