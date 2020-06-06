#!/bin/bash

if [[ $# -eq 0 ]]; then
  N_CONNECTIONS=1         # Number of connections done per client
  N_CLIENTS=1             # Number of clients running on one vm
  N_VMS=1                 # Total number of vm's in this test (1, 2 or 3)
elif [[ $# -eq 1 ]]; then
  N_CONNECTIONS=$1
  N_CLIENTS=1
  N_VMS=1
else
  N_CONNECTIONS=$1
  N_CLIENTS=$2
  N_VMS=1
fi

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
  printf "Cannot start test, unknown host %d.\n" "$HOST"
  exit 1
fi

printf "Start test..\n"
printf "VM ID: %d, #VMs: %d, #Clients: %d, #Connections: %d\n\n" "$HOST_VM_ID" "$N_VMS" "$N_CLIENTS" "$N_CONNECTIONS"

# Path for the output
DATE=$(date +%F-%H-%M-%S)
OUTPUT_PATH=output/"$DATE"/vm$HOST_VM_ID

# Start fresh with iperf
pkill iperf

# Start the server listening on the given ports

for PORT in "${PORTS[@]}"; do
  mkdir -p "$OUTPUT_PATH"/iperf/server/
  ip netns exec shila-ingress iperf3 -s -p "$PORT" > "$OUTPUT_PATH"/iperf/server/"$PORT".log 2> "$OUTPUT_PATH"/iperf/server/"$PORT".err &
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

REPORT_OUTPUT_FILE=report.txt
touch "$REPORT_OUTPUT_FILE"

printf "\n++++ REPORT ++++\n" >  "$REPORT_OUTPUT_FILE"
printf "VM ID: %d, #VMs: %d, #Clients: %d, #Connections: %d\n\n" "$HOST_VM_ID" "$N_VMS" "$N_CLIENTS" "$N_CONNECTIONS" >> "$REPORT_OUTPUT_FILE"

for OUTPUT_FILE in $(find "$OUTPUT_PATH" -type f -name iperf-client-*); do
  cat "$OUTPUT_FILE" >> "$REPORT_OUTPUT_FILE"
done

printf "Done.\n"
exit 0