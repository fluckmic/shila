#!/bin/bash

clear

ADDRESS="10.7.0.9"

SRC_ID=0
DST_ID=1
TRANSFER_VALUE=100
INTERVAL=1

SEND_FOR_DURATION=0
SEND_FOR_SIZE=1


if [[ $# -eq 1 ]]; then
  TRANSFER_VALUE=$1
elif [[ $# -eq 2 ]]; then
  SRC_ID=$1
  DST_ID=$2
elif [[ $# -eq 3 ]]; then
  SRC_ID=$1
  DST_ID=$2
  TRANSFER_VALUE=$3
fi

mapfile -t PORTS < IperfListeningPorts.data
mapfile -t CLIENTS < hostNames.data

SRC_CLIENT=${CLIENTS["$SRC_ID"]}
DST_CLIENT=${CLIENTS["$DST_ID"]}

PORT=${PORTS["$DST_ID"]}

if [[ SEND_FOR_DURATION -eq 1 ]]; then
  printf "Send for %s seconds from Client %d to Client %d (port %s).\n" "$TRANSFER_VALUE" "$SRC_ID" "$DST_ID" "$PORT"
  CMD="sudo ip netns exec shila-egress iperf3 -c ""$ADDRESS"" -p ""$PORT"" -t ""$TRANSFER_VALUE"" -i ""$INTERVAL --get-server-output -M 500"
  echo "$CMD"
  sshpass -f client.password ssh -tt scion@"$SRC_CLIENT" -q "$CMD" 2>&1 | tee -a "_iperfClient""$SRC_ID"".log"
    if [[ $? -ne 0 ]]; then
      printf "Failure : Unable to send.\n"
      exit 1
    fi
fi

if [[ SEND_FOR_SIZE -eq 1 ]]; then

  N_BUFFER_WRITE_CYCLES=$(($TRANSFER_VALUE * 10))

  printf "Send %s MByte from Client %d to Client %d (port %s).\n" "$TRANSFER_VALUE" "$SRC_ID" "$DST_ID" "$PORT"
  CMD="sudo ip netns exec shila-egress iperf3 -c ""$ADDRESS"" -p ""$PORT"" -l 100K -n ""$N_BUFFER_WRITE_CYCLES""  -i ""$INTERVAL --get-server-output -M 500"
  echo "$CMD"
  sshpass -f client.password ssh -tt scion@"$SRC_CLIENT" -q "$CMD" 2>&1 | tee -a "_iperfClient""$SRC_ID"".log"
    if [[ $? -ne 0 ]]; then
      printf "Failure : Unable to send.\n"
      exit 1
    fi
fi

