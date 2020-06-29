#!/bin/bash

clear

PATH_TO_EXPERIMENT="~/go/src/shila/testing/MultiInstance"

CLIENT=$1

if   [[ "$CLIENT" == "mptcp-over-scion-vm-0" ]]; then
  CONFIG_FILE="config0.json"
elif [[ "$CLIENT" == "mptcp-over-scion-vm-1" ]]; then
  CONFIG_FILE="config1.json"
else
  printf "Failure : " "Unknown host %s.\n" "$CLIENT"
  exit 1
fi

# Initialize
SCRIPT_NAME="init"
CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to initialize %s.\n" "$CLIENT"
  exit 1
 fi

CMD="./_shila -config ""$CONFIG_FILE"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to initialize %s.\n" "$CLIENT"
  exit 1
 fi