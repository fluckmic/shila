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
elif [[ $# -eq 4 ]]; then
  SRC_ID=$1
  DST_ID=$2
  DURATION=$3
  FOR_REPORT=$4
fi

mapfile -t PORTS < iperfListeningPorts.data
PORT=${PORTS["$DST_ID"]}


if [[ $FOR_REPORT -eq 1 ]]; then
	PORT=27041
fi

NAMESPACE="shila-egress-""$SRC_ID"

clear
if [[ $FOR_REPORT -eq 0 ]]; then
	printf "Send for %s seconds from Client %s to Client %s (port %s).\n" "$DURATION" "$SRC_ID" "$DST_ID" "$PORT"
	CMD="iperf3 -c ""$ADDRESS"" -p ""$PORT"" -t ""$DURATION"" -i ""$INTERVAL"" --get-server-output"
	echo "$CMD"
fi
if [[ $FOR_REPORT -eq 1 ]]; then 
	printf "Host %d - Starting iPerf3..\n\n" $(($SRC_ID + 1))
	CMD="iperf3 -c ""$ADDRESS"" -p ""$PORT"" -t ""$DURATION"" -i ""$INTERVAL"
fi

sudo ip netns exec "$NAMESPACE" $CMD
