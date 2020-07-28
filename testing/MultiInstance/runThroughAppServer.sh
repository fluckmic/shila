#!/bin/bash

clear

PATH_TO_EXPERIMENT="~/go/src/shila/testing/MultiInstance"

CLIENT_ID=$1

CLIENT="mptcp-over-scion-vm-""$CLIENT_ID"

if   [[ $CLIENT_ID -eq 0 ]]; then
  PORT=60000
elif [[ $CLIENT_ID -eq 1 ]]; then
  PORT=60001
elif [[ $CLIENT_ID -eq 2 ]]; then
  PORT=60002
  elif [[ $CLIENT_ID -eq 3 ]]; then
  PORT=60003
else
  printf "Failure : " "Unknown Client %d.\n" "$CLIENT_ID"
  exit 1
fi

printf "Client %d - Starting ThroughApp..\n\n" "$CLIENT_ID"
sleep 15

CMD="cd ""$PATH_TO_EXPERIMENT""sudo ip netns exec shila-ingress ./_throughApp -s -p""$PORT"" -r"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD" 2>&1 | tee -a "_rThroughAppServer""$CLIENT_ID"".log"
if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to start receiving ThroughApp server on Client %d.\n" "$CLIENT_ID"
  exit 1
fi
