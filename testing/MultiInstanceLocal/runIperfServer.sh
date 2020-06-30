#!/bin/bash

clear
sleep 15

CLIENT_ID=$1

mapfile -t PORTS < iperfListeningPorts.data
mapfile -t INGRESS_NAMESPACES < ingressNamespaces.data

PORT=${PORTS["$CLIENT_ID"]}
NAMESPACE=${INGRESS_NAMESPACES["$CLIENT_ID"]}

ip netns exec "$NAMESPACE" iperf3 -s -p "$PORT"
