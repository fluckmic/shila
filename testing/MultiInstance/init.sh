#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

# Kill running instances
pkill shila

# Delete all namespaces
bash ../../helper/netnsClear.sh

## Update the repo
git pull

## Build the latest version of all software required
export PATH=$PATH:/usr/local/go/bin
# shila
go build -o _shila ../../

# Restart SCION
sudo systemctl stop scionlab.target
sleep 2
sudo systemctl start scionlab.target
sleep 2