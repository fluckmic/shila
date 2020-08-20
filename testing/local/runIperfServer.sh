#!/bin/bash

clear

CLIENT_ID=$1

mapfile -t PORTS < iperfListeningPorts.data

PORT=${PORTS["$CLIENT_ID"]}
NAMESPACE="shila-ingress-""$CLIENT_ID"

printf "Client %d - Starting iperf3..\n\n" "$CLIENT_ID"

sleep 15
ip netns exec "$NAMESPACE" iperf3 -s -p "$PORT"
