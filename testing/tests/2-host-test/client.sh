#!/usr/bin/env bash

HOST_VM_ID="$1"    # ID number of this vm
N_VMS="$2"         # Total number of vm's in this test
CLIENT_ID="$3"     # ID of the client
N_CONNECTIONS="$4" # Number of connections per client

MAX_DURATION=20

PORTS=(2727 4411 6688 7321 8686)

echo ""
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

  echo Client "$CLIENT_ID" on mptcp"$HOST_VM_ID" tries to connect to mptcp"$TARGET_VM_ID" on port "$PORT".
  iperf -c mptcp"$TARGET_VM_ID" -p "$PORT" -t "$DURATION"

  sleep 2
done
