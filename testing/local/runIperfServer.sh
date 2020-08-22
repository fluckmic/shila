#!/bin/bash

clear

CLIENT_ID=$1
FOR_REPORT=$2

mapfile -t PORTS < iperfListeningPorts.data

PORT=${PORTS["$CLIENT_ID"]}
NAMESPACE="shila-ingress-""$CLIENT_ID"

if [[ $FOR_REPORT -eq 0 ]]; then
	printf "Client %d - Starting iperf3..\n\n" "$CLIENT_ID"
fi

if [[ $FOR_REPORT -eq 1 ]]; then
	printf "Host %d - Starting iPerf3..\n\n" $(($CLIENT_ID + 1))
	PORT=27041
fi

sleep 15
ip netns exec "$NAMESPACE" iperf3 -s -p "$PORT"
