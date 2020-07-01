#!/bin/bash

LOGFILE_EXPERIMENT=$3

if [[ $2 -eq 1 ]]; then
  printf "Debug : %s\n" "$1" | tee -a "$LOGFILE_EXPERIMENT"
fi