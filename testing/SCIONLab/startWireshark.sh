#!/bin/bash

# Starts a wireshark instance in the egress and the ingress namespace of a client.

SRC_ID=$1
DST_ID=$2

SRC_CLIENT="mptcp-over-scion-vm-""$SRC_ID"
DST_CLIENT="mptcp-over-scion-vm-""$DST_ID"

clear

printf "Starting wireshark instances egress on Client %d and ingress on Client %d.\n" "$SRC_ID" "$DST_ID"

ssh scion@"$SRC_CLIENT" "sudo ip netns exec shila-egress sudo tcpdump -s 0 -w -" | wireshark -k -i - &
ssh scion@"$DST_CLIENT" "sudo ip netns exec shila-ingress sudo tcpdump -s 0 -w -" | wireshark -k -i - &
