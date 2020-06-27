#!/bin/bash

echo "Enter terminate session."

# Terminates a tmux session (if it exits).
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

echo "After base change"

SESSION_NAME=$1

./isRunningSession.sh "$SESSION_NAME"
if [[ $? -eq 0 ]]; then
  tmux kill-session -t "$SESSION_NAME" &>/dev/null
fi

exit 0