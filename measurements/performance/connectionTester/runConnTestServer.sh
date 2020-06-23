#!/bin/bash

HOST_NAME=$(uname -n)

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

printf "Starting connection test server %s.\n" "$HOST_NAME"
../._connTest -name "$HOST_NAME" -p 27271

exit $!

