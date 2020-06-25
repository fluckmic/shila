#!/bin/bash
SESSION=$USER

bash ../../helper/netnsClear.sh

tmux kill-session

tmux -2 new-session -d -s $SESSION

tmux new-window -t $SESSION:1 

tmux split-window -h
tmux select-pane -t 0
tmux split-window -v
tmux select-pane -t 1
tmux split-window -v
tmux select-pane -t 3
tmux split-window -v
tmux split-window -v

tmux select-pane -t 0
tmux send-keys "sudo bash runShila1.sh" C-m
tmux select-pane -t 2
tmux send-keys "sudo bash runServer1.sh" C-m
tmux select-pane -t 3
tmux send-keys "sudo bash runShila2.sh" C-m
tmux select-pane -t 5
tmux send-keys "sudo bash runServer2.sh" C-m
tmux select-pane -t 1
tmux send-keys "clear" C-m
tmux select-pane -t 4
tmux send-keys "clear" C-m

# Attach to session
tmux -2 attach-session -t $SESSION
