#!/bin/bash

## Load the hosts name and the base directory
HOST_NAME=$(uname -n)

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

printf "Starting initialization of %s.\n" "$HOST_NAME"

##Determine the hosts id and make it available
if   [[ "$HOST_NAME" == "mptcp-over-scion-vm-1" ]]; then
  HOST_ID=1
elif [[ "$HOST_NAME" == "mptcp-over-scion-vm-2l" ]]; then
  HOST_ID=2
elif [[ "$HOST_NAME" == "mptcp-over-scion-vm-3l" ]]; then
  HOST_ID=3
elif [[ "$HOST_NAME" == "mptcp-over-scion-vm-4l" ]]; then
  HOST_ID=4
else
  printf "Failed - Cannot determine host id.\n"
  exit 1
fi
echo "$HOST_ID" > _HOST_ID

## Update the repo
git pull

## Build the latest version of all software required
export PATH=$PATH:/usr/local/go/bin

# shila
go build ../../


printf "Initialization of %s done.\n" "$HOST_NAME"
exit 0

