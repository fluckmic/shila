#!/bin/bash

# Starts a new detached tmux session.
tmux new-session -d -s "$@"
exit 0