#!/bin/bash

## Load the hosts name and the base directory
HOST_NAME=$(uname -n)
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

printf "Starting initialization of %s.\n" "$HOST_NAME" > _init.log

## Remove all unnecessary stuff.
rm -f _*
#./cleanUp.sh 1

## Determine the hosts id and make it available
if   [[ "$HOST_NAME" == "mptcp-over-scion-vm-0" ]]; then
  HOST_ID=0
elif [[ "$HOST_NAME" == "mptcp-over-scion-vm-1" ]]; then
  HOST_ID=1
elif [[ "$HOST_NAME" == "mptcp-over-scion-vm-2" ]]; then
  HOST_ID=2
elif [[ "$HOST_NAME" == "mptcp-over-scion-vm-3" ]]; then
  HOST_ID=3
else
  printf "Initialization of %s failed - Cannot determine host id.\n" "$HOST_NAME" > _init.err
  exit 1
fi
echo "$HOST_ID" > _hostId
echo "$HOST_NAME" > _hostName

## Update the repo
git pull

## Build the latest version of all software required
export PATH=$PATH:/usr/local/go/bin

# Connection tester
go build -o _connTest ./connectionTester

# shila
go build -o _shila ../../

printf "Initialization of %s done.\n" "$HOST_NAME" > _init.log
exit 0

