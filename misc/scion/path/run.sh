#!/bin/bash

# Run the server with
# ./run.sh s

# Run the client with
# ./run.sh c

if [[ $1 == c ]]; then
    printf "Run as client.\n"
    export SCION_DAEMON_ADDRESS=127.0.0.44:30255
    go run path.go -remote 2-ff00:0:220,[127.0.0.1]:2727
fi

if [[ $1 == s ]]; then
    printf "Run as server.\n"
    export SCION_DAEMON_ADDRESS=127.0.0.68:30255
    go run path.go -port 2727
fi


#   In  the shila.topo topology
