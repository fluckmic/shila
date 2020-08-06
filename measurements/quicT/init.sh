#!/bin/bash

## Load the hosts name and the base directory
HOST_NAME=$(uname -n)
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

printf "Starting initialization of %s.\n" "$HOST_NAME" > _init.log

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

## Build the latest version of all software required
export PATH=$PATH:/usr/local/go/bin

cp quicT.go ~/go/src/github.com/scionproto/scion/go/examples/pingpong/quicT.go
cd ~/go/src/github.com/scionproto/scion/go/examples/pingpong/
go build -o _quicT quicT.go
cd ~/go/src/shila/measurements/quicT/
cp ~/go/src/github.com/scionproto/scion/go/examples/pingpong/_quicT .

sleep 1

printf "Initialization of %s done.\n" "$HOST_NAME" > _init.log

exit 0
