#!/bin/bash

clear

CLIENT_ID=$1

mapfile -t PORTS < iperfListeningPorts.data

PORT=${PORTS["$CLIENT_ID"]}
NAMESPACE="shila-ingress-""$CLIENT_ID"

printf "Client %d - Starting server instance for delay measurement..\n\n" "$CLIENT_ID"

gcc ../../measurements/delay/server.c -o _delayMeasurementServer

sleep 15
ip netns exec "$NAMESPACE" ./_delayMeasurementServer -p "$PORT" -d
