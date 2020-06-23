#!/bin/bash

HOST_NAME=$(uname -n)

printf "Starting connection test server %s.\n" "$HOST_NAME"
../._connTest -name "$HOST_NAME" -p 27271

return $!

