#!/bin/bash

clear
sleep 15

CLIENT=$1

if   [[ "$CLIENT" == "mptcp-over-scion-vm-0" ]]; then
  PORT=60000
elif [[ "$CLIENT" == "mptcp-over-scion-vm-1" ]]; then
  PORT=60001
elif [[ "$CLIENT" == "mptcp-over-scion-vm-2" ]]; then
  PORT=60002
  elif [[ "$CLIENT" == "mptcp-over-scion-vm-3" ]]; then
  PORT=60003
else
  printf "Failure : " "Unknown host %s.\n" "$CLIENT"
  exit 1
fi

printf "Client %d - Starting iperf3..\n\n" "$CLIENT"

CMD="sudo ip netns exec shila-ingress iperf3 -s -p""$PORT"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD"
if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to start iperf server on %s.\n" "$CLIENT"
  exit 1
fi
