#!/bin/bash

clear

printf "Starting a single client..\n"
sudo ip netns exec shila-egress iperf -c 10.7.0.9 -p 2727
