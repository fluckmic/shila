#!/bin/bash

# Terminates a tmux session (if it exits).
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

SESSION_NAME=$1

echo "$SESSION_NAME"

./isRunningSession.sh "$SESSION_NAME"
if [[ $? -eq 0 ]]; then
  echo "Before kill-session"
  tmux kill-session -t "$SESSION_NAME" &>/dev/null
fi

exit 0