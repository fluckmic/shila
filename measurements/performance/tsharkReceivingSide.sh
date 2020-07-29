#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_tsharkReceivingSide.log"
ERR_FILE="_tsharkReceivingSide.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

TSHARK_DUMP_FILENAME="_incomingTraffic.pcap"

touch "$TSHARK_DUMP_FILENAME"
sudo chmod o=rw

sleep 1

printf "Starting tshark on the receiving side %s.\n" "$HOST_NAME" >> "$LOG_FILE"
tshark -i eth0 -w "$TSHARK_DUMP_FILENAME" -F pcap > "$LOG_FILE" 2> "$ERR_FILE"

exit 0