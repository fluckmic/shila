#!/bin/bash

clear

export PATH=$PATH:/usr/local/go/bin

cd ../../
go build

# Shila4.topo
#export SCION_DAEMON_ADDRESS=127.0.0.44:30255

# Tiny.topo
export SCION_DAEMON_ADDRESS=127.0.0.19:30255

./shila -config testing/MultiInstanceLocal/config1.json