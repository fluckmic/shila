#!/bin/bash

CONNECTION_TEST_TIMEOUT=15

PATH_TO_EXPERIMENT=$(dirname "$0")

EXPERIMENT_NAME="Performance measurement"

CLIENTS=(mptcp-over-scion-vm-0 mptcp-over-scion-vm-1 mptcp-over-scion-vm-2)
CLIENT_IDS=(0 1 2)
N_REPETITIONS=5
N_INTERFACES=(1 2 3 7 10)
PATH_SELECTIONS=(0 1)
DURATION=60

DURATION_BETWEEN=10

N_EXPERIMENTS=$((${#CLIENTS[@]} * (${#CLIENTS[@]} - 1) * $N_REPETITIONS * ${#N_INTERFACES[@]} * ${#PATH_SELECTIONS[@]}))
TOTAL_DURATION=$(((($DURATION + $DURATION_BETWEEN) * $N_EXPERIMENTS) / 3600 ))

START_SESSION="bash ~/go/src/shila/measurements/sessionScripts/startSession.sh"
CHECK_ERROR="bash ~/go/src/shila/measurements/sessionScripts/checkForError.sh"
KILL_ALL_SESSIONS="bash ~/go/src/shila/measurements/sessionScripts/terminateAllSessions.sh"

clear
########################################################################################################################
## Do a simple test to see if we are able to establish a connection

printf "Starting %s:\n\n" "$EXPERIMENT_NAME"
printf "Clients:\t"; echo "${CLIENTS[@]}"
printf "Interfaces:\t"; echo "${N_INTERFACES[@]}"
printf "Path selection:\t"; echo "${PATH_SELECTIONS[@]}"
printf "Duration:\t%s\n" "$DURATION"
printf "\nTotal number of experiments:\t%d\n" "$N_EXPERIMENTS"
printf "Estimated duration:\t\t%dh\n\n" "$TOTAL_DURATION"
########################################################################################################################
## Create the experiments file

rm -f _experiments.data

for SRC in "${CLIENT_IDS[@]}"; do
  for DST in "${CLIENT_IDS[@]}"; do
  if [[ $SRC != $DST ]]; then
    for N_INTERFACE in "${N_INTERFACES[@]}"; do
      for PATH_SELECT in "${PATH_SELECTIONS[@]}"; do
        COUNT=1
        while [[ "$COUNT" -le "$N_REPETITIONS" ]]; do
          echo "$SRC" "$DST" "$N_INTERFACE" "$PATH_SELECT" "$COUNT" >> _experiments.data
          COUNT=$(($COUNT+1))
        done
      done
    done
  fi
  done
done

# Create a random order
shuf _experiments.data | shuf -o _experiments.data


















