#!/bin/bash

echo "Hi"

# Starts a new detached tmux session.
tmux new-session -d -s "$@"
exit 0