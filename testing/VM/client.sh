#!/bin/bash

HOST_VM_ID="$1"    # ID number of this vm
N_VMS="$2"         # Total number of vm's in this test (1, 2 or 3)
CLIENT_ID="$3"     # ID of the client
N_CONNECTIONS="$4" # Number of connections per client
OUTPUT_PATH="$5"   # Were to write the output data

MAX_DURATION=5

# Select the ports
if [[ "$HOST_VM_ID" -eq 1 ]]; then
  if [[ "$N_VMS" -eq 1 ]]; then
    PORTS=(11111 11112 11113 11114)
  elif [[ "$N_VMS" -eq 2 ]]; then
    PORTS=(11111 11112 11113 11114 22221 22222 22223 22224)
  else
    PORTS=(11111 11112 11113 11114 22221 22222 22223 22224 33331 33332 33333 33334)
  fi
elif [[ "$HOST_VM_ID" -eq 2 ]]; then
  if [[ "$N_VMS" -eq 1 ]]; then
    PORTS=(22221 22222 22223 22224)
  elif [[ "$N_VMS" -eq 2 ]]; then
    PORTS=(11111 11112 11113 11114 22221 22222 22223 22224)
  else
    PORTS=(11111 11112 11113 11114 22221 22222 22223 22224 33331 33332 33333 33334)
  fi
elif [[ "$HOST_VM_ID" -eq 3 ]]; then
  if [[ "$N_VMS" -eq 1 ]]; then
    PORTS=(33331 33332 33333 33334)
  elif [[ "$N_VMS" -eq 2 ]]; then
    PORTS=(33331 33332 33333 33334 22221 22222 22223 22224)
  else
    PORTS=(11111 11112 11113 11114 22221 22222 22223 22224 33331 33332 33333 33334)
  fi
else
  exit 1
fi

mkdir -p "$OUTPUT_PATH"/iperf/client/"$CLIENT_ID"
CLIENT_OUTPUT_PATH="$OUTPUT_PATH"/iperf-client-"$CLIENT_ID".log

printf "\n++++ CLIENT LOG ++++\n" >> "$CLIENT_OUTPUT_PATH"
printf "(vm id: %d) (client id: %d) (connections: %d) (max duration: %d)\n\n" "$HOST_VM_ID" "$CLIENT_ID" "$N_CONNECTIONS" "$MAX_DURATION"  >> "$CLIENT_OUTPUT_PATH"

for (( CONN=0; CONN<"$N_CONNECTIONS"; CONN++ ))
do
  # Select a port at random.
  PORT="${PORTS[(($(( RANDOM % ${#PORTS[@]} ))))]}"

  # Select the duration of the connection at random
  DURATION=$(( RANDOM % MAX_DURATION + 1 ))

  printf "+ Connection to 10.7.0.9:%d:\n" "$PORT" >> "$CLIENT_OUTPUT_PATH"
  ip netns exec shila-egress iperf3 -c 10.7.0.9 -p "$PORT" -t "$DURATION" >> "$CLIENT_OUTPUT_PATH" 2>> "$CLIENT_OUTPUT_PATH"

  wait $!
done

echo "$CLIENT_LOG_HEADER4" >> "$CLIENT_OUTPUT_PATH"

printf "Client %d done.\n" "$CLIENT_ID"
exit 0