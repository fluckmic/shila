#!/bin/bash

clear

sudo ip netns exec shila-egress-1 iperf3 -c 10.7.0.9 -p 22222 -t 20 -i 1 -P 1 -V
