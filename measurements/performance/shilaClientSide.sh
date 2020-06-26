#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

LOG_FILE="_shilaClientSide.log"
ERR_FILE="_shilaClientSide.err"

HOST_NAME=$(cat _hostName)
HOST_ID=$(cat _hostId)

PATH_SELECTIONS=("\"mtu\"" "\"shortest\"")

N_VIRTUAL_INTERFACES=$1
PATH_SELECTION="${PATH_SELECTIONS[$2]}"

sed "s/@1/""$N_VIRTUAL_INTERFACES""/g" config.data | sed "s/@2/""$PATH_SELECTION""/g" > _clientConfig.json

printf "Starting shila on the client side %s.\n" "$HOST_NAME" >> "$LOG_FILE"
sudo ./_shila -config _clientConfig.json >> "$LOG_FILE" 2>> "$ERR_FILE"