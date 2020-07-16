#!/bin/bash

clear

ADDRESS="10.7.0.9"

SRC_ID=0
DST_ID=1
DURATION=100
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

mapfile -t PORTS < IperfListeningPorts.data
mapfile -t CLIENTS < hostNames.data

SRC_CLIENT=${CLIENTS["$SRC_ID"]}
DST_CLIENT=${CLIENTS["$DST_ID"]}

PORT=${PORTS["$DST_ID"]}

printf "Send for %s seconds from Client %d to Client %d (port %s).\n" "$DURATION" "$SRC_ID" "$DST_ID" "$PORT"

CMD="sudo ip netns exec shila-egress iperf3 -c ""$ADDRESS"" -p ""$PORT"" -t ""$DURATION"" -i ""$INTERVAL --get-server-output -M 500"
echo "$CMD"
sshpass -f client.password ssh -tt scion@"$SRC_CLIENT" -q "$CMD" 2>&1 | tee -a "_iperfClient""$SRC_ID"".log"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to send.\n"
  exit 1
 fi
