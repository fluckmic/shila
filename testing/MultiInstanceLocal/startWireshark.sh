#!/bin/bash

sudo ip netns exec shila-egress-1 wireshark &
sudo ip netns exec shila-ingress-2 wireshark &
