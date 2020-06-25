#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_shilaServerSide.log"
ERR_FILE="_shilaServerSide.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

printf "Starting shila on the server side %s.\n" "$HOST_NAME" >> "$LOG_FILE"
