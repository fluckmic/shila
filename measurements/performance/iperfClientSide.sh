#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

rm -f _iperfClientSide*

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

ADDRESS="10.7.0.9"

REMOTE_ID=$1
N_INTERFACE=$2
PATH_SELECT=$3
REPETITION=$4

LOG_FILE="_iperfClientSide_""$HOST_ID""_""$REMOTE_ID""_""$N_INTERFACE""_""$PATH_SELECT""_""$REPETITION"".log"
ERR_FILE="_iperfClientSide.err"

DURATION=120
INTERVAL=2

mapfile -t PORTS < IperfListeningPorts.data
PORT=${PORTS["$REMOTE_ID"]}

printf "Starting iperf on the client side %s for %ds (%ds interval).\n" "$HOST_NAME" "$DURATION" "$INTERVAL" >> "$LOG_FILE"
printf "%s %s %s %s %s %s %s\n" "$HOST_ID" "$REMOTE_ID" "$N_INTERFACE" "$PATH_SELECT" "$REPETITION" "$DURATION" "$INTERVAL" >> "$LOG_FILE"
iperf -c "$ADDRESS" -p "$PORT" -t "$DURATION" -i "$INTERVAL" >> "$LOG_FILE" 2>> "$ERR_FILE"