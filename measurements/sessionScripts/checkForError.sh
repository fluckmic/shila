#!/bin/bash

# Checks whether a certain session returned with an error.

SESSION_NAME=$1
PATH_TO_EXPERIMENT=$2

ERROR_FILE="$PATH_TO_EXPERIMENT""/_""$SESSION_NAME"".err"

if [[ -f "$ERROR_FILE" ]]; then
  printf "There is an error file %s.\n" "$ERROR_FILE"
  ERROR_FILE_SIZE=$(du "$ERROR_FILE" | cut -f1)
  if [[ "$ERROR_FILE_SIZE" -gt 0 ]]; then
    cat "$ERROR_FILE"
    exit 1
  fi
fi

exit 0