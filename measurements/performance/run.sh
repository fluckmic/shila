#!/bin/bash

CLIENTS=(vm-1 vm-2 vm-3 vm-4)

## First initialize all clients
for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q 'sudo bash ~/go/src/shila/measurements/performance/init.sh'
  if [[ $? -ne 0 ]]; then
    printf "Failed to initialize %s.\n" "$CLIENT"
    exit 1
  fi
done

## Then do a connection checks.
#  Start the connection test servers.
for SERVER in "${CLIENTS[@]}"; do
  ssh -tt scion@"$SERVER" -q 'sudo bash ~/go/src/shila/measurements/performance/connectionTester/runConnTestServer.sh'
  if [[  $? -ne 0 ]]; then
    printf "Failed to start connection test server %s.\n" "$SERVER"
    exit 1
  fi
done

#  Run the client side.
for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q 'sudo bash ~/go/src/shila/measurements/performance/connectionTester/runConnTestClient.sh'
  if [[  $? -ne 0 ]]; then
    printf "Connection test for client %s failed.\n" "$CLIENT"
    exit 1
  fi
done

