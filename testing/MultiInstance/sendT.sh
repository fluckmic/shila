#!/bin/bash

clear

PATH_TO_EXPERIMENT="~/go/src/shila/testing/MultiInstance"

ADDRESS="10.7.0.9"

SRC_ID=0
DST_ID=1
DATA_TRANSFER=10
INTERVAL=1

if [[ $# -eq 1 ]]; then
  DATA_TRANSFER=$1
elif [[ $# -eq 2 ]]; then
  SRC_ID=$1
  DST_ID=$2
elif [[ $# -eq 3 ]]; then
  SRC_ID=$1
  DST_ID=$2
  DATA_TRANSFER=$3
fi

mapfile -t PORTS < IperfListeningPorts.data
mapfile -t CLIENTS < hostNames.data

SRC_CLIENT=${CLIENTS["$SRC_ID"]}
DST_CLIENT=${CLIENTS["$DST_ID"]}

PORT=${PORTS["$DST_ID"]}

printf "Send %s MBytes from Client %d to Client %d (port %s).\n" "$DATA_TRANSFER" "$SRC_ID" "$DST_ID" "$PORT"

CMD="cd ""$PATH_TO_EXPERIMENT""; sudo ip netns exec shila-egress ./_throughApp -c ""$ADDRESS"" -p ""$PORT"" -n ""$DURATION"
echo "$CMD"
sshpass -f client.password ssh -tt scion@"$SRC_CLIENT" -q "$CMD" 2>&1 | tee -a "_sThroughAppClient""$SRC_ID"".log"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to send.\n"
  exit 1
 fi
