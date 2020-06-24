#!/bin/bash

# Terminates a tmux session (if it exits).

SESSION_NAME=$1

tmux kill-session -t "$SESSION_NAME" 2>/dev/null