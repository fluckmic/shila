#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_tsharkSCIONTrafficPost.log"
ERR_FILE="_tsharkSCIONTrafficPost.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

SCION_TRAFFIC_PCAP_DUMP_FILENAME="_tsharkSCIONTraffic.pcap"
SCION_TRAFFIC_CSV_DUMP_FILENAME="_tsharkSCIONTraffic.csv"

printf "Starting post processing captured SCION traffic on %s.\n" "$HOST_NAME" >> "$LOG_FILE"
tshark -r "$SCION_TRAFFIC_PCAP_DUMP_FILENAME" -T fields -e frame.number -e frame.time -e udp.length -E separator=, -E header=y > "$SCION_TRAFFIC_CSV_DUMP_FILENAME" 2> "$ERR_FILE"

sleep 1