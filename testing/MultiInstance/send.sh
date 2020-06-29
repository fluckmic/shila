#!/bin/bash

ADDRESS="10.7.0.9"

SRC_ID=1
DST_ID=0
DURATION=10
INTERVAL=1

mapfile -t PORTS < IperfListeningPorts.data
mapfile -t CLIENTS < hostNames.data

SRC_CLIENT=${CLIENTS["$SRC_ID"]}
DST_CLIENT=${CLIENTS["$DST_ID"]}

PORT=${PORTS["$DST_ID"]}

printf "Send for %s seconds from %s to %s (port %s).\n" "$DURATION" "$SRC_CLIENT" "$DST_CLIENT" "$PORT"

CMD="sudo ip netns exec shila-egress iperf -c ""$ADDRESS"" -p ""$PORT"" -t ""$DURATION"" -i ""$INTERVAL"
echo "$CMD"
sshpass -f client.password ssh -tt scion@"$SRC_CLIENT" -q "$CMD"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to send.\n"
  exit 1
 fi
