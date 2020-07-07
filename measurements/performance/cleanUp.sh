#!/bin/bash

## Load the hosts name and the base directory
HOST_NAME=$(uname -n)
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

## Reset everything to be ready for a new repetition.

pkill shila
pkill iperf

sleep 5

# Delete all namespaces
bash ../../helper/netnsClear.sh

# Remove all log and error files.
rm -f _*.log
rm -f _*.err

# Remove all builds as well.
if [[ $1 -eq 1 ]]; then
  rm -f _*
fi

exit 0