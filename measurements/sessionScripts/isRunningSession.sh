#!/bin/bash

# Checks whether a session with a certain name is up and running.

SESSION_NAME=$1

tmux has-session -t "$SESSION_NAME" 2>/dev/null
exit 0