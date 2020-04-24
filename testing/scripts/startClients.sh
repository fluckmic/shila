#!/bin/bash

clear
nOfClients=4
echo "Starting" $nOfClients "clients."
cnt=1
while [ $cnt -le $nOfClients ]
do
  ip netns exec shila-egress iperf -c 10.7.0.9 > log-client-$cnt.txt 2>&1 &
  ((cnt++))
done
