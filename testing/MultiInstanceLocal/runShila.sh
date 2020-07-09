#!/bin/bash

clear

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

export PATH=$PATH:/usr/local/go/bin

CLIENT_ID=$1
CONFIG_FILE="config""$CLIENT_ID"".json"

go build -o _shila ../../

mapfile -t DAEMONS < daemonAddresses.data
DAEMON=${DAEMONS["$CLIENT_ID"]}
export SCION_DAEMON_ADDRESS="$DAEMON"
./_shila -config "$CONFIG_FILE" | tee -a "_""$CLIENT_ID""_output.log"
