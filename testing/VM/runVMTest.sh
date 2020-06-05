#!/bin/bash

N_VMS=2           # Total number of vm's in this test
N_CLIENTS=3       # Number of clients running on one vm
N_CONNECTIONS=4   # Number of connections done per client

# Load the host id
HOST=$(uname -n)
if   [[ "$HOST" == "mptcp-over-scion-vm-1" ]]; then
  HOST_VM_ID=1
  PORTS=(11111 11112 11113 11114)
elif [[ "$HOST" == "mptcp-over-scion-vm-2" ]]; then
  HOST_VM_ID=2
  PORTS=(22221 22222 22223 22224)
elif [[ "$HOST" == "mptcp-over-scion-vm-3" ]]; then
  HOST_VM_ID=3
  PORTS=(33331 33332 33333 33334)
else
  printf "Cannot start test, unknown host %d." "$HOST"
  exit 1
fi

echo ""
echo "$HOST" \(id:"$HOST_VM_ID"\)
echo "$N_CLIENTS" client\(s\) with "$N_CONNECTIONS" connections each
echo ""

# Update the repo
git pull

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

# Build the latest version
/usr/local/go/bin/go build ../../

mkdir -p "$OUTPUT_PATH"/shila/

touch "$OUTPUT_PATH"/shila/shila.log
printf   "++++ SHILA LOG  ++++\n" >> "$OUTPUT_PATH"/shila/shila.log

../.././shila >> "$OUTPUT_PATH"/shila/shila.log 2>> "$OUTPUT_PATH"/shila/shila.log &
printf "Started shila..\n"

# Start fresh with iperf
pkill iperf

# Start the server listening on the given ports

for PORT in "${PORTS[@]}"; do
  mkdir -p "$OUTPUT_PATH"/iperf/server/
  iperf3 -s -p "$PORT" > "$OUTPUT_PATH"/iperf/server/"$PORT".log 2> "$OUTPUT_PATH"/iperf/server/"$PORT".err &
  printf "Started iperf server listening on port %d.\n" "$PORT"
done

# Start the clients
CLIENT_PIDS=()
for (( CLIENT_ID=0; CLIENT_ID<"$N_CLIENTS"; CLIENT_ID++ ))
do
  ./client.sh "$HOST_VM_ID" "$N_VMS" "$CLIENT_ID" "$N_CONNECTIONS" "$OUTPUT_PATH" &
  CLIENT_PIDS+=($!)
done

printf "\nStarted the clients, waiting for them to finish..\n"

for CLIENT_PID in "${CLIENT_PIDS[@]}"; do
  wait "$CLIENT_PID"
done

printf "\nAll clients done.\nCreate report..\n"

touch report.txt
for OUTPUT_FILE in $(find "$OUTPUT_PATH" -type f -name shila.*); do
  cat "$OUTPUT_FILE" >> report.txt
done
for OUTPUT_FILE in $(find "$OUTPUT_PATH" -type f -name iperf-client-*); do
  cat "$OUTPUT_FILE" >> report.txt
done

printf "Done.\n"
exit 0