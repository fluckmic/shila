#!/bin/bash

export PATH=$PATH:/usr/local/go/bin

cd ../../
go build
export SCION_DAEMON_ADDRESS=127.0.0.68:30255
./shila -config testing/MultiInstanceLocal/config2.json
