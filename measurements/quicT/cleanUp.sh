#!/bin/bash

## Load the hosts name and the base directory
HOST_NAME=$(uname -n)
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

## Reset everything to be ready for a new repetition.

pkill tshark
pkill quicT

sleep 1

# Remove all log, error, dump and config files.
rm -f _*.log
rm -f _*.err
rm -f _*.csv
rm -f _*.pcap

# Remove all builds as well.
if [[ $1 -eq 1 ]]; then
  rm -f _*
fi

exit 0