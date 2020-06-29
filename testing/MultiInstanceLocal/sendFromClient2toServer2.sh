#!/bin/bash
clear

sudo ip netns exec shila-egress-2 iperf -c 10.7.0.9 -p 22222 -t 20 -i 1
