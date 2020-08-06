#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

REMOTE_ID=$1
DIRECTION=$2
REPETITION=$3
TRANSFER=$4

if [[ $DIRECTION -eq 1 ]]; then
  LOG_FILE="_quicTReceiverSide.log"
  ERR_FILE="_quicTReceiverSide.err"
fi
if [[ $DIRECTION -eq 0 ]]; then
  LOG_FILE="_quicTSenderSide.log"
  ERR_FILE="_quicTSenderSide.err"
fi

mapfile -t ADDRESSES < scionAddresses.data

SRC_ADDR="${ADDRESSES["$HOST_ID"]}"
DST_ADDR="${ADDRESSES["$REMOTE_ID"]}"

printf "Starting quicT on the client side %s.\n" "$HOST_NAME" >> "$LOG_FILE"
printf "HostID RemoteID Direction Repetition Transfer\n"  >> "$LOG_FILE"
printf "%s %s %s %s %s\n" "$HOST_ID" "$REMOTE_ID" "$DIRECTION" "$REPETITION" "$TRANSFER"  >> "$LOG_FILE"

if [[ $DIRECTION -eq 1 ]]; then
  ./_quicT -mode client -remote "$DST_ADDR" -local "$SRC_ADDR" -n "$TRANSFER" -R >> "$LOG_FILE" 2>> "$ERR_FILE"
fi

if [[ $DIRECTION -eq 0 ]]; then
  ./_quicT -mode client -remote "$DST_ADDR" -local "$SRC_ADDR" -n "$TRANSFER" >> "$LOG_FILE" 2>> "$ERR_FILE"
fi

exit 0
