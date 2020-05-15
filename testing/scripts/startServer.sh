#!/bin/bash

clear

printf "Starting iperf server listening at port 2727.. \n"
sudo ip netns exec shila-ingress iperf -s -p 2727