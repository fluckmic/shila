#!/bin/bash

clear
sleep 5
ip netns exec shila-ingress-1 iperf -s -p 11113
