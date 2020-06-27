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
#rm -f _*.log
#rm -f _*.err

# Kill all tmux sessions
#sudo pkill -f tmux

.././terminateSession.sh "shilaClientSide"
.././terminateSession.sh "shilaServerSide"
.././terminateSession.sh "iperfClientSide"
.././terminateSession.sh "iperfServerSide"

# Kill all shila and iperf instances
#sudo pkill shila &>/dev/null
#sudo pkill iperf &>/dev/null

# Delete all namespaces
bash ../../helper/netnsClear.sh

exit 0