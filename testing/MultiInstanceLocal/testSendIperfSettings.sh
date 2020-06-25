#!/bin/bash

clear

ADDRESS="10.7.0.9"
PORT=11113

DURATION=120
INTERVAL=2

sudo ip netns exec shila-egress-1 iperf -c "$ADDRESS" -p "$PORT" -t  "$DURATION" -i "$INTERVAL" >>  test.log 2>> test.err
