#!/bin/bash

clear

PATH_TO_EXPERIMENT="~/go/src/shila/testing/MultiInstance"

CLIENT=$1

if   [[ "$CLIENT" == "mptcp-over-scion-vm-0" ]]; then
  CONFIG_FILE="config0.json"
elif [[ "$CLIENT" == "mptcp-over-scion-vm-1" ]]; then
  CONFIG_FILE="config1.json"
elif [[ "$CLIENT" == "mptcp-over-scion-vm-2" ]]; then
  CONFIG_FILE="config2.json"
elif [[ "$CLIENT" == "mptcp-over-scion-vm-3" ]]; then
  CONFIG_FILE="config3.json"
else
  exit 1
fi

# Initialize
SCRIPT_NAME="init"
CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
echo "$CMD"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD"
#ssh -tt scion@"$CLIENT" -q "$CMD"
RET=$?
 if [[ $RET -ne 0 ]]; then
  printf "Failure : Unable to initialize %s. Error code: %d.\n" "$CLIENT" "$RET"
  exit 1
 fi

printf "%s\n\n" "$CLIENT"

CMD="cd ""$PATH_TO_EXPERIMENT""; sudo ./_shila -config ""$CONFIG_FILE"
sshpass -f client.password ssh -tt scion@"$CLIENT" -q "$CMD" | tee -a "_""$CLIENT""_output.log"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to initialize %s.\n" "$CLIENT"
  exit 1
 fi