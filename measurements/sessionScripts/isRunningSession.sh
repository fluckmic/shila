#!/bin/bash

# Checks whether a session with a certain name is up and running.

SESSION_NAME=$1

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

printf "Check if there is a running session %s?\n" "$SESSION_NAME"

tmux has-session -t "$SESSION_NAME" 2>/dev/null
if [[ $? -eq 0 ]]; then
  printf "There is a running session %s.\n" "$SESSION_NAME"
  exit 0
else
  printf "There is no running session %s.\n" "$SESSION_NAME"
  exit 1
fi
