#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

rm -f _iperfServerSide*

LOG_FILE="_iperfServerSide.log"
ERR_FILE="_iperfServerSide.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

mapfile -t PORTS < IperfListeningPorts.data
PORT="${PORTS["$HOST_ID"]}"

printf "Starting iperf on the server side %s listening on port %s.\n" "$HOST_NAME" "$PORT" >> "$LOG_FILE"
iperf -s -p "$PORT" >> "$LOG_FILE" 2>> "$ERR_FILE"
