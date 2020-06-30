#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

# Kill running instances
pkill shila
pkill iperf

# Delete all namespaces
bash ../../helper/netnsClear.sh

## Update the repo
git pull &>/dev/null

## Build the latest version of all software required
export PATH=$PATH:/usr/local/go/bin
# shila
go build -o _shila ../../

sudo systemctl restart scionlab.target
sleep 2