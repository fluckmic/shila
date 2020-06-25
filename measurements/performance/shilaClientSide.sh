#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_shilaClientSide.log"
ERR_FILE="_shilaClientSide.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

printf "Starting shila on the client side %s.\n" "$HOST_NAME" >> "$LOG_FILE"