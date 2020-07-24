#!/bin/bash

DEST_IP="10.7.0.9"

SRC_ID=0
DST_ID=1
N_WRITES=10
INTERVAL=1

if [[ $# -eq 1 ]]; then
  DURATION=$1
elif [[ $# -eq 2 ]]; then
  SRC_ID=$1
  DST_ID=$2
elif [[ $# -eq 3 ]]; then
  SRC_ID=$1
  DST_ID=$2
  N_WRITES=$3
fi

mapfile -t PORTS < iperfListeningPorts.data
PORT=${PORTS["$DST_ID"]}

ADDITIONAL_LOG_INFO="Client ""$SRC_ID"

NAMESPACE="shila-egress-""$SRC_ID"

gcc ../../measurements/delay/client.c -o _delayMeasurementClient

clear
sudo ip netns exec "$NAMESPACE" ./_delayMeasurementClient -c "$DEST_IP" -p "$PORT" -f "_client""$SRC_ID""EgressTimestamps.log" -n "$N_WRITES" -a "$ADDITIONAL_LOG_INFO" -d
