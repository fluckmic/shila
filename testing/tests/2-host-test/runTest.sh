#!/usr/bin/env bash

N_VMS=2            # Total number of vm's in this test
N_CLIENTS=1        # Number of clients running on one vm
N_CONNECTIONS=5    # Number of connections done per client

# Load test id (name of the folder)
TEST=${PWD##*/}

# Load host id
HOST=$(uname -n)

if   [[ "$HOST" == "mptcp-over-scion-vm-1" ]]; then
  ID=1
elif [[ "$HOST" == "mptcp-over-scion-vm-2" ]]; then
  ID=2
elif [[ "$HOST" == "mptcp-over-scion-vm-3" ]]; then
  ID=3
else
  echo Cannot start test, unknown host "$HOST".
  exit 1
fi

# Reset the output folders
rm -f -d -r output/
mkdir -p output/$ID/stdio output/$ID/stderr

# Copy the routing file such that it is found by shila
cp routing$ID.json ../../../

# Start shila
pkill shila

../../.././shila >> output/$ID/stdio/shila 2>> output/$ID/stderr/shila &
PID_SHILA=$!

# Run the test
./test.sh "$ID" "$N_VMS" "$N_CLIENTS" "$N_CONNECTIONS"
