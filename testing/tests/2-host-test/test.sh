#!/usr/bin/env bash

TEST_ID="2HostTest"

HOST_VM_ID="$1"    # ID number of this vm
N_VMS="$2"         # Total number of vm's in this test
N_CLIENTS="$3"     # Number of clients running on this vm
N_CONNECTIONS="$4" # Number of connections per client

# Start fresh
pkill iperf

# Start the server listening on the given ports
PORTS=(2727 4411 6688 7321 8686)
for PORT in "${PORTS[@]}"; do
  iperf -s -p "PORT" >> output/"$HOST_VM_ID"/stdio/iperfServer"$PORT" 2>> output/"$HOST_VM_ID"/stderr/iperfServer"$PORT" &
done

for (( CLIENT_ID=0; CLIENT_ID<"$N_CLIENTS"; CLIENT_ID++ ))
do
  ./client.sh "$HOST_VM_ID" "$N_VMS" "$CLIENT_ID" "$N_CONNECTIONS" &
done