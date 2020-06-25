#!/bin/bash

ip netns list | cut -f1 --delimiter=" " > tmp.txt
mapfile -t INTERFACES < tmp.txt

for INTERFACE in "${INTERFACES[@]}"; do
  sudo ip netns delete "$INTERFACE"
done

rm tmp.txt