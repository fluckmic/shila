#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_tsharkMPTCPTrafficPost.log"
ERR_FILE="_tsharkMPTCPTrafficPost.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

MPTCP_TRAFFIC_PCAP_DUMP_FILENAME="_tsharkMPTCPTraffic.pcap"
MPTCP_TRAFFIC_CSV_DUMP_FILENAME="_tsharkMPTCPTraffic.csv"

printf "Starting post processing captured MPTCP traffic on %s.\n" "$HOST_NAME" >> "$LOG_FILE"
#tshark -r "$MPTCP_TRAFFIC_PCAP_DUMP_FILENAME" -T fields -e frame.number -e frame.time -e udp.length -E separator=, -E header=y > "$MPTCP_TRAFFIC_CSV_DUMP_FILENAME" 2> "$ERR_FILE"

sleep 1