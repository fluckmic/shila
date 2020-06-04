#!/usr/bin/env bash

HOST=$(uname -n)

if   [[ "$HOST" == "mptcp-over-scion-vm-1" ]]; then
  ID=1
elif [[ "$HOST" == "mptcp-over-scion-vm-2" ]]; then
  ID=2
elif [[ "$HOST" == "mptcp-over-scion-vm-3" ]]; then
  ID=3
else
  echo "Cannot start test, unknown host $HOST."
fi

cp "routing$ID.json" ../../../
