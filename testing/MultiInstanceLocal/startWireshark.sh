#!/bin/bash

# Starts a wireshark instance in the egress and the ingress namespace of a client.

CLIENT_ID=$1

INGRESS_NAMESPACE="shila-ingress-""$CLIENT_ID"
EGRESS_NAMESPACE="shila-egress-""$CLIENT_ID"

clear

printf "Starting wireshark instances in namespaces of Client %d.\n" "$CLIENT_ID"

sudo ip netns exec "$EGRESS_NAMESPACE" wireshark &
sudo ip netns exec "$INGRESS_NAMESPACE" wireshark &
