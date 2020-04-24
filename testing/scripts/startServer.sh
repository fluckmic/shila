#!/bin/bash

clear
echo "Starting server."
ip netns exec shila-ingress iperf -s