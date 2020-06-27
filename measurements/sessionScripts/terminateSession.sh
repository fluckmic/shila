#!/bin/bash

echo "Enter terminate session."

# Terminates a tmux session (if it exits).
BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

SESSION_NAME=$1

./isRunningSession.sh "$SESSION_NAME"
RETURN=$?
echo "After isRunningSession"
if [[ "$RETURN" -eq 0 ]]; then
  tmux kill-session -t "$SESSION_NAME" &>/dev/null
fi

exit 0