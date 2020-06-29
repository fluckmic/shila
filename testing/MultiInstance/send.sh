#!/bin/bash

ADDRESS="10.7.0.9"

SRC_ID=0
DST_ID=1
DURATION=10
INTERVAL=1


mapfile -t PORTS < IperfListeningPorts.data
mapfile -t CLIENTS < hostNames.data

SRC_CLIENT=${CLIENTS["$SRC_ID"]}
DST_CLIENT=${CLIENTS["$DST_ID"]}

PORT=${PORTS["$REMOTE_ID"]}

printf "Send for %s seconds from %s to %s.\n" "$DURATION" "$SRC_CLIENT" "$DST_CLIENT"

#sudo ip netns exec shila-egress iperf -c "$ADDRESS" -p "$PORT" -t "$DURATION" -i "$INTERVAL"
