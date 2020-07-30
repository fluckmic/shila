#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_tsharkMPTCPTraffic.log"
ERR_FILE="_tsharkMPTCPTraffic.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

MPTCP_TRAFFIC_PCAP_DUMP_FILENAME="_tsharkMPTCPTraffic.pcap"

touch "$MPTCP_TRAFFIC_PCAP_DUMP_FILENAME"
sudo chmod o=rw "$MPTCP_TRAFFIC_PCAP_DUMP_FILENAME"

CAPTURE_FILTER=""

printf "Starting capturing MPTCP traffic on the receiving side %s.\n" "$HOST_NAME" >> "$LOG_FILE"
sudo ip netns exec shila-ingress tshark -i tun1 -w "$MPTCP_TRAFFIC_PCAP_DUMP_FILENAME" -F pcap > "$LOG_FILE" 2> "$ERR_FILE"

sleep 1