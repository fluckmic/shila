#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

ADDRESS="10.7.0.9"

REMOTE_ID=$1; N_INTERFACE=$2; PATH_SELECT=$3; REPETITION=$4; DURATION=$5
INTERVAL=1

LOG_FILE="_iperfClientSide_""$HOST_ID""_""$REMOTE_ID""_""$N_INTERFACE""_""$PATH_SELECT""_""$REPETITION"".log"
ERR_FILE="_iperfClientSide.err"

mapfile -t PORTS < IperfListeningPorts.data
PORT=${PORTS["$REMOTE_ID"]}

printf "Starting iperf on the client side %s for %ds (%ds interval).\n" "$HOST_NAME" "$DURATION" "$INTERVAL" >> "$LOG_FILE"
printf "HostID RemoteID Address Port Repetition PathSelect Duration Interval\n" >> "$LOG_FILE"
printf "%s %s %s %s %s %s %s %s.\n" "$HOST_ID" "$REMOTE_ID" "$ADDRESS" "$PORT" "$REPETITION" "$PATH_SELECT" "$DURATION" "$INTERVAL" >> "$LOG_FILE"
sudo ip netns exec shila-egress iperf -c "$ADDRESS" -p "$PORT" -t "$DURATION" -i "$INTERVAL" >> "$LOG_FILE" 2>> "$ERR_FILE"