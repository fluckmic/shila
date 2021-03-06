#!/bin/bash
SESSION=$USER

rm -f _*

#bash ../../helper/netnsClear.sh

tmux kill-session

tmux -2 new-session -d -s $SESSION

tmux new-window -t $SESSION:1 

tmux split-window -v
tmux select-pane -t 0
tmux split-window -h
tmux select-pane -t 2
tmux split-window -v
tmux split-window -h
tmux select-pane -t 2

tmux select-pane -t 0
tmux send-keys "sudo bash runShila.sh 0" C-m
tmux select-pane -t 1
tmux send-keys "sudo bash runShila.sh 1" C-m
tmux select-pane -t 3
tmux send-keys "sudo bash runIperfServer.sh 0" C-m
tmux select-pane -t 4
tmux send-keys "sudo bash runIperfServer.sh 1" C-m

tmux select-pane -t 2

# Attach to session
tmux -2 attach-session -t $SESSION
