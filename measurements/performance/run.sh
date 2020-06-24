#!/bin/bash

CLIENTS=(vm-1 vm-2)

START_SESSION='bash ~/go/src/shila/measurements/sessionScripts/startSession.sh'
CHECK_SESSION='bash ~/go/src/shila/measurements/sessionScripts/isRunningSession.sh'

## First initialize all clients
for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q "$START_SESSION" "init" "sudo bash ~/go/src/shila/measurements/performance/init.sh"
done


for CLIENT in "${CLIENTS[@]}"; do
  RUNNING=0
  while [ "$RUNNING" -eq 0  ]; do
      ssh -tt scion@"$CLIENT" -q "$CHECK_SESSION" "init"
      RUNNING=$?
  done
  printf "Client %s is done with %s.\n" "init"
done


## Then do a connection checks.
#  Start the connection test servers.
#for SERVER in "${CLIENTS[@]}"; do
#  ssh -tt scion@"$SERVER" -q 'nohup sudo bash ~/go/src/shila/measurements/performance/connectionTester/runConnTestServer.sh >> ConnTestServer.log 2>&1 &'
#  if [[  $? -ne 0 ]]; then
#    printf "Failed to start connection test server %s.\n" "$SERVER"
#    exit 1
#  fi
#done

#  Run the client side.
#for CLIENT in "${CLIENTS[@]}"; do
#  ssh -tt scion@"$CLIENT" -q 'sudo bash ~/go/src/shila/measurements/performance/connectionTester/runConnTestClient.sh'
#  if [[  $? -ne 0 ]]; then
#    printf "Connection test for client %s failed.\n" "$CLIENT"
#    exit 1
#  fi
#done

