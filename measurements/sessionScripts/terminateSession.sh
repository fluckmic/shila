#!/bin/bash

# Terminates a tmux session (if it exits).

SESSION_NAME=$1

./isRunningSession.sh "$SESSION_NAME"
if [[ $? -eq 0 ]]; then
  tmux kill-session -t "$SESSION_NAME" 2>/dev/null
fi

exit 0