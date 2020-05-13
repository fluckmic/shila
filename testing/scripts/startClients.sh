#!/bin/bash

clear

nOfClients=4
printf "Starting %d clients..\n" $nOfClients

cnt=1
while [ $cnt -le $nOfClients ]
do
  ip netns exec shila-egress iperf -c 10.7.0.9 -p 2727 > log-client-$cnt.txt 2>&1 &
  printf "Started client %d. \n" $cnt
  ((cnt++))
done
