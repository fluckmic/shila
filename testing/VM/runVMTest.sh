#!/bin/bash

N_VMS=2            # Total number of vm's in this test
N_CLIENTS=1        # Number of clients running on one vm
N_CONNECTIONS=5    # Number of connections done per client

# Load the host id
HOST=$(uname -n)
if   [[ "$HOST" == "mptcp-over-scion-vm-1" ]]; then
  HOST_VM_ID=1
elif [[ "$HOST" == "mptcp-over-scion-vm-2" ]]; then
  HOST_VM_ID=2
elif [[ "$HOST" == "mptcp-over-scion-vm-3" ]]; then
  HOST_VM_ID=3
else
  echo Cannot start test, unknown host "$HOST".
  exit 1
fi

echo ""
echo "$HOST" \(id:"$HOST_VM_ID"\)
echo "$N_CLIENTS" client\(s\) with "$N_CONNECTIONS" connections each
echo ""

# Return if vm is not meant to be part of the test
if [[ "$HOST_VM_ID" -gt "$N_VMS" ]]; then
  exit 0
fi

# Path for the output
DATE=$(date +%F-%H-%M-%S)
OUTPUT_PATH=output/"$DATE"/vm$HOST_VM_ID

# Setup and run shila
pkill shila
# Copy the routing file such that it is found by shila
cp routing$HOST_VM_ID.json ../../

mkdir -p "$OUTPUT_PATH"/shila/
../.././shila > "$OUTPUT_PATH"/shila/shila.log 2> "$OUTPUT_PATH"/shila/shila.err &
echo Started shila..
echo ""

# Start fresh with iperf
pkill iperf

# Start the server listening on the given ports
PORTS=(2727 4411 6688 7321 8686)
for PORT in "${PORTS[@]}"; do
  mkdir -p "$OUTPUT_PATH"/iperf/server/
  iperf -s -p "$PORT" > "$OUTPUT_PATH"/iperf/server/"$PORT".log 2> "$OUTPUT_PATH"/iperf/server/"$PORT".err &
  echo Started iperf server listening on port "$PORT"..
done
echo ""

# Start the clients
for (( CLIENT_ID=0; CLIENT_ID<"$N_CLIENTS"; CLIENT_ID++ ))
do
  ./client.sh "$HOST_VM_ID" "$N_VMS" "$CLIENT_ID" "$N_CONNECTIONS" "$OUTPUT_PATH" &
done
echo Started the clients, waiting for them to finish..

exit 0