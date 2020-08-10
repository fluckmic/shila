#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

DURATION_MODE=1
TRANSFER_MODE=2

MSS=1024

ADDRESS="10.7.0.9"

REMOTE_ID=$1; DIRECTION=$2; N_INTERFACE=$3; PATH_SELECT=$4; REPETITION=$5; VALUE=$6; MODE=$7;
INTERVAL=1

LOG_FILE="_iperfClientSide_""$HOST_ID""_""$REMOTE_ID""_""$DIRECTION""_""$N_INTERFACE""_""$PATH_SELECT""_""$REPETITION"".log"
ERR_FILE="_iperfClientSide.err"

mapfile -t PORTS < IperfListeningPorts.data
PORT=${PORTS["$REMOTE_ID"]}

if [[ $MODE -eq $DURATION_MODE ]]; then
  DYN_ARG="-t ""$VALUE"" -i ""$INTERVAL"
  VALUE_DESC="Duration"
fi
if [[ $MODE -eq $TRANSFER_MODE ]]; then
  DYN_ARG="-n ""$VALUE"
  VALUE_DESC="Transfer"
fi
if [[ $DIRECTION -eq 1 ]]; then
  DYN_ARG="$DYN_ARG"" -R"
fi

printf "HostID RemoteID Address Port Repetition PathSelect %s Interval nInterfaces Direction\n" "$VALUE_DESC" >> "$LOG_FILE"
printf "%s %s %s %s %s %s %s %s %s %s.\n" "$HOST_ID" "$REMOTE_ID" "$ADDRESS" "$PORT" "$REPETITION" "$PATH_SELECT" "$VALUE" "$INTERVAL" "$N_INTERFACE" "$DIRECTION" >> "$LOG_FILE"
CMD="iperf3 -c ""$ADDRESS"" -p ""$PORT"" --get-server-output -M ""$MSS"" ""$DYN_ARG"
sudo ip netns exec shila-egress $CMD >> "$LOG_FILE" 2>> "$ERR_FILE"

exit 0
