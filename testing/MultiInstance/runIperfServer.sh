#!/bin/bash

clear
sleep 10

CLIENT=$1

if   [[ "$CLIENT" == "mptcp-over-scion-vm-0" ]]; then
  PORT=11111
elif [[ "$CLIENT" == "mptcp-over-scion-vm-1" ]]; then
  PORT=22222
else
  printf "Failure : " "Unknown host %s.\n" "$CLIENT"
  exit 1
fi

CMD="sudo ip netns exec shila-ingress iperf3 -s -p""$PORT"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD"
if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to start iperf server on %s.\n" "$CLIENT"
  exit 1
fi
