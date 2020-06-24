#!/bin/bash

clear
sleep 5
ip netns exec shila-ingress-2 iperf -s -p 22222
