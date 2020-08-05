#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

REMOTE_ID=$1
DIRECTION=$2
REPETITION=$3
TRANSFER=$4

if [[ $DIRECTION -eq 1 ]]; then
  LOG_FILE="_quicTSenderSide.log"
  ERR_FILE="_quicTSenderSide.err"
fi

if [[ $DIRECTION -eq 0 ]]; then
  LOG_FILE="_quicTReceiverSide.log"
  ERR_FILE="_quicTReceiverSide.err"
fi

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

mapfile -t ADDRESSES < scionAddresses.data

SRC_ADDR="${ADDRESSES["$HOST_ID"]}"

printf "Starting quicT on the server side %s.\n" "$HOST_NAME" >> "$LOG_FILE"
printf "HostID RemoteID Repetition Direction Transfer\n"  >> "$LOG_FILE"
printf "%s %s %s %s %s\n" "$HOST_ID" "$REMOTE_ID" "$REPETITION" "$TRANSFER" "$DIRECTION"

if [[ $DIRECTION -eq 1 ]]; then
  ./quicT -mode server -local "$SRC_ADDR" -R >> "$LOG_FILE" 2>> "$ERR_FILE"
fi

if [[ $DIRECTION -eq 0 ]]; then
  ./quicT -mode server -local "$SRC_ADDR" >> "$LOG_FILE" 2>> "$ERR_FILE"
fi
