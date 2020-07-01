#!/bin/bash

ADDRESS="10.7.0.9"

SRC_ID=0
DST_ID=1
DURATION=10
INTERVAL=1

if [[ $# -eq 1 ]]; then
  DURATION=$1
elif [[ $# -eq 2 ]]; then
  SRC_ID=$1
  DST_ID=$2
elif [[ $# -eq 3 ]]; then
  SRC_ID=$1
  DST_ID=$2
  DURATION=$3
fi

mapfile -t PORTS < iperfListeningPorts.data
PORT=${PORTS["$DST_ID"]}

mapfile -t EGRESS_NAMESPACES < egressNamespaces.data
EGRESS_NAMESPACE=${EGRESS_NAMESPACES["$SRC_ID"]}

printf "Send for %s seconds from client %s to client %s (port %s).\n" "$DURATION" "$SRC_ID" "$DST_ID" "$PORT"
CMD="iperf3 -c ""$ADDRESS"" -p ""$PORT"" -t ""$DURATION"" -i ""$INTERVAL"" --get-server-output"
echo "$CMD"
sudo ip netns exec "$EGRESS_NAMESPACE" $CMD
