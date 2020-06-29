#!/bin/bash

clear
sleep 10
ip netns exec shila-ingress-1 iperf3 -s -p 11113
