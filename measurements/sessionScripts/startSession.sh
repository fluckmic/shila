#!/bin/bash

# Starts a new detached tmux session.

SESSION_NAME=$1
SESSION_CMD=$2

echo "$@"

printf "Starting new session %s with command %s.\n" "$SESSION_NAME" "$SESSION_CMD"
tmux new-session -d -s "$SESSION_NAME" "$SESSION_CMD"