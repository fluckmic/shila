#!/bin/bash

CLIENTS=(vm-1 vm-2 vm-3 vm-4)

# First initialize all clients
for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q 'sudo bash ~/go/src/shila/measurements/performance/init.sh'
  if [[ ! $? ]]; then
    printf "Failed to initialize %s.\n" "$CLIENT"
    exit 1
  fi
done

# Then do a connection check.


