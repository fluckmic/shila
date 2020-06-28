#!/bin/bash

# Checks whether a session with a certain name is up and running.

SESSION_NAME=$1

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

tmux has-session -t "$SESSION_NAME" 2>/dev/null
if [[ $? -eq 0 ]]; then
  exit 0
else
  exit 1
fi
