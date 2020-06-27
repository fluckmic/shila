#!/bin/bash

# Checks whether a session with a certain name is up and running.

SESSION_NAME=$1

echo "in is running session"
echo "$SESSION_NAME"

tmux has-session -t "$SESSION_NAME" &>/dev/null
RETURN=$?

echo "$RETURN"

exit "$RETURN"