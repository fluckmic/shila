#!/bin/bash

CLIENTS=(vm-1 vm-2 vm-3 vm-4)
CLIENT_ADDR=(17-ffaa:1:d87 19-ffaa:1:d88 20-ffaa:1:d89 18-ffaa:1:d8a)

# First initialize all clients
for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q 'sudo bash ~/go/src/shila/measurements/performance/init.sh'
  if [[ $? -ne 0 ]]; then
    printf "Failed to initialize %s.\n" "$CLIENT"
    exit 1
  fi
done

# Then do a connection checks.
for SERVER in "${CLIENTS[@]}"; do

  ssh -tt scion@"$SERVER" -q 'sudo bash ~/go/src/shila/measurements/performance/connectionTester/runConnTestServer.sh'
    echo $?
    if [[  $? -ne 0 ]]; then
    printf "Failed to start connection test server %s.\n" "$SERVER"
    exit 1
    fi

    for CLIENT in "${CLIENTS[@]}"; do
      if [[ "$CLIENT" == "$SERVER" ]]; then
          continue
      fi

    done

 done

