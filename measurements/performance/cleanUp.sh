#!/bin/bash

## Load the hosts name and the base directory
HOST_NAME=$(uname -n)
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

## Reset everything to be ready for a new repetition.

echo "Start cleanup."

# Remove all builds as well.
if [[ $1 -eq 1 ]]; then
  rm -f _*
fi

echo "1"

# Remove all log and error files.
rm -f _*.log
rm -f _*.err

echo "2"

# Kill all tmux sessions
tmux kill-server 2>/dev/null

echo "3"

# Kill all shila and iperf instances
sudo pkill shila
sudo pkill iperf

echo "4"

# Delete all namespaces
bash ../../helper/netnsClear.sh

echo "Done cleanup."

exit 0