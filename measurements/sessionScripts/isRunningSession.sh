#!/bin/bash

# Checks whether a session with a certain name is up and running.

SESSION_NAME=$1

echo "In isRunningSession"

tmux has-session -t "$SESSION_NAME" &>/dev/null

