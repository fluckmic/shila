#!/bin/bash

clear

PATH_TO_EXPERIMENT="~/go/src/shila/testing/MultiInstance"

CLIENT_ID=$1

CONFIG_FILE="config""$CLIENT_ID"".json"
CLIENT="mptcp-over-scion-vm-""$CLIENT_ID"

# Initialize
SCRIPT_NAME="init"
CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
echo "$CMD"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD"
#ssh -tt scion@"$CLIENT" -q "$CMD"
RET=$?
 if [[ $RET -ne 0 ]]; then
  printf "Failure : Unable to initialize Client %d. Error code: %d.\n" "$CLIENT_ID" "$RET"
  exit 1
 fi

printf "Client %d - Starting shila..\n\n" "$CLIENT_ID" | tee -a "_shila""$CLIENT_ID"".log"

CMD="cd ""$PATH_TO_EXPERIMENT""; sudo ./_shila -config ""$CONFIG_FILE"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD" 2>&1 | tee -a "_shila""$CLIENT_ID"".log"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to initialize Client %d.\n" "$CLIENT_ID"
  exit 1
 fi