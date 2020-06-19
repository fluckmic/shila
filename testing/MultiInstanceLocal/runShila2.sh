#!/bin/bash

clear

export PATH=$PATH:/usr/local/go/bin

cd ../../
go build

# Shila4.topo
# export SCION_DAEMON_ADDRESS=127.0.0.68:30255

# Tiny.topo
export SCION_DAEMON_ADDRESS=[fd00:f00d:cafe::7f00:b]:30255

./shila -config testing/MultiInstanceLocal/config2.json
