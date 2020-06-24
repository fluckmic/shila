#!/bin/bash

# Checks whether a certain session returned with an error.

SESSION_NAME=$1
PATH_TO_EXPERIMENT=$2

ERROR_FILE="$PATH_TO_EXPERIMENT""/_""$SESSION_NAME"

if [[ -f "$ERROR_FILE" ]]; then
  cat "$ERROR_FILE"
  exit 1
fi

exit 0
