#!/bin/bash

HOST_VM_ID="$1"    # ID number of this vm
N_VMS="$2"         # Total number of vm's in this test
CLIENT_ID="$3"     # ID of the client
N_CONNECTIONS="$4" # Number of connections per client
OUTPUT_PATH="$5"   # Were to write the output data

MAX_DURATION=20

PORTS=(2727 4411 6688 7321 8686)

mkdir -p "$OUTPUT_PATH"/iperf/client/

for (( CONN=0; CONN<"$N_CONNECTIONS"; CONN++ ))
do
  # Select a port at random.
  PORT="${PORTS[(($(( RANDOM % ${#PORTS[@]} ))))]}"

  # Select the duration of the connection at random
  DURATION=$(( RANDOM % MAX_DURATION + 1 ))

  # Select the target vm
  TARGET_VM_ID=$(( RANDOM % N_VMS + 1 ))
  while [[ TARGET_VM_ID -eq HOST_VM_ID ]]; do
    TARGET_VM_ID=$(( RANDOM % N_VMS + 1 ))
    done

  iperf -c mptcp"$TARGET_VM_ID" -p "$PORT" -t "$DURATION" \
    >> "$OUTPUT_PATH"/iperf/client/"$CLIENT_ID"-"$PORT".log \
    2>> "$OUTPUT_PATH"/iperf/client/"$CLIENT_ID"-"$PORT".err &

  sleep 2
done
printf "Client %d done.\n" "$CLIENT_ID"
exit 0