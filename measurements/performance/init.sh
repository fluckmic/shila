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

CONGESTION_CONTROL=$1

## Update the repo
git pull > _init.log 2> _init.err
if [[ $? -ne 0 ]]; then
  printf "Initialization of %s failed - Issue with git.\n" "$HOST_NAME" > _init.err
  exit 1
fi

## Build the latest version of all software required
export PATH=$PATH:/usr/local/go/bin
# shila
go build -o _shila ../../

sleep 1

# Configure MPTCP and the congestion control algorithm
sudo sysctl net.mptcp.mptcp_scheduler=roundrobin
sudo sysctl net.mptcp.mptcp_path_manager=fullmesh

sudo sysctl net.ipv4.tcp_congestion_control="$CONGESTION_CONTROL"

printf "Initialization of %s done.\n" "$HOST_NAME" > _init.log
exit 0

