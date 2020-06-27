#!/bin/bash

## Load the hosts name and the base directory
HOST_NAME=$(uname -n)
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

## Reset everything to be ready for a new repetition.

# Remove all builds as well.
if [[ $1 -eq 1 ]]; then
  rm -f _*
fi

# Remove all log and error files.
rm -f _*.log
rm -f _*.err

# Kill all tmux sessions
tmux kill-server 2>/dev/null

# Kill all shila and iperf instances
#sudo pkill shila >&2>/dev/null
sudo pkill iperf >&2>/dev/null

# Delete all namespaces
bash ../../helper/netnsClear.sh

exit 0