#!/bin/bash

# Setup and run a local scion network with the Tiny4.topo
# ./scion topology -c topology/Tiny4.topo
# ./scion run

# If necessary figure out ip's of scion daemon ad edit
# cat gen/ISD1/ASff00_0_110/endhost/sd.toml
# cat gen/ISD1/ASff00_0_111/endhost/sd.toml
# cat gen/ISD1/ASff00_0_112/endhost/sd.toml

# Compile

# Run the server with
# ./run.sh s

# Run the client with
# ./run.sh c

if [[ $1 == c ]]; then
    printf "Run as client.\n"
    # export SCION_DAEMON_ADDRESS=127.0.0.19:30255
#    go run intro.go -remote 17-ffaa:1:d87,127.0.0.1:2727
	go run intro.go -remote 19-ffaa:1:d88,[127.0.0.1]:27271
fi

if [[ $1 == s ]]; then
    printf "Run as server.\n"

    # Tiny4.topo
    # export SCION_DAEMON_ADDRESS=127.0.0.19:30255

    # Tiny.topo
    # export SCION_DAEMON_ADDRESS=[fd00:f00d:cafe::7f00:b]:30255

    go run intro.go -port 27271
fi


#   sd: 127.0.0.12:30255              sd: 127.0.0.19:30255
#   [1-ff00:0:110]--------------------[1-ff00:0:111]
#         |
#         |
#         |
#   [1-ff00:0:112]
#   sd: 127.0.0.27:30255
