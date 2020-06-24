#!/bin/bash

# Starts a new detached tmux session.

SESSION_NAME=$1
SESSION_CMD=$2

tmux new-session -d -s "$SESSION_NAME" "$SESSION_CMD"