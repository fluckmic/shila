#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_tsharkSCIONTraffic.log"
ERR_FILE="_tsharkSCIONTraffic.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

TSHARK_DUMP_FILENAME="_tsharkSCIONTraffic.pcap"

CAPTURE_FILTER="udp dst port 50000"

printf "Starting tshark on the receiving side %s.\n" "$HOST_NAME" >> "$LOG_FILE"
tshark -i eth0 -f "$CAPTURE_FILTER" -w "$TSHARK_DUMP_FILENAME" -F pcap > "$LOG_FILE" 2> "$ERR_FILE"

sleep 1