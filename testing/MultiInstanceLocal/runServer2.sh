#!/bin/bash

clear
sleep 10
ip netns exec shila-ingress-2 iperf -s -p 22222
