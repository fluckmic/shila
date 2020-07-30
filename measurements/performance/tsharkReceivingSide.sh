#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_tsharkReceivingSide.log"
ERR_FILE="_tsharkReceivingSide.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

TSHARK_DUMP_FILENAME="_incomingTraffic.pcap"

CAPTURE_FILTER="udp dst port 50000"

#touch "$TSHARK_DUMP_FILENAME"
#sudo chmod o=rw "$TSHARK_DUMP_FILENAME"

printf "Starting tshark on the receiving side %s.\n" "$HOST_NAME" >> "$LOG_FILE"
tshark -i eth0 -f "$CAPTURE_FILTER" -w "$TSHARK_DUMP_FILENAME" -F pcap > "$LOG_FILE" 2> "$ERR_FILE"

sleep 1