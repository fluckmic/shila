#!/bin/bash
SESSION=$USER

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
tmux select-pane -t 1
tmux send-keys "echo 1" C-m
tmux select-pane -t 2
tmux send-keys "sudo bash runServer1.sh" C-m
tmux select-pane -t 3
tmux send-keys "sudo bash runShila2.sh" C-m
tmux select-pane -t 4
tmux send-keys "echo 4" C-m
tmux select-pane -t 5
tmux send-keys "sudo bash runServer2.sh" C-m

#tmux send-keys "ls" C-m
#tmux send-keys "ls" C-m

# Setup a MySQL window
#tmux new-window -t $SESSION:2 -n 'MySQL' 'mysql -uroot'

# Set default window
#tmux select-window -t $SESSION:1

# Attach to session
tmux -2 attach-session -t $SESSION
