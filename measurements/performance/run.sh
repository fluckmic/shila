#!/bin/bash

CONNECTION_TEST_TIMEOUT=15

PATH_TO_EXPERIMENT=$(dirname "$0")

EXPERIMENT_NAME="Performance measurement"

CLIENTS=(mptcp-over-scion-vm-0 mptcp-over-scion-vm-1 mptcp-over-scion-vm-2)
CLIENT_IDS=(0 1 2)
N_REPETITIONS=5
N_INTERFACES=(1 2 3 7 10)
PATH_SELECTION=(0 1)
DURATION=60

DURATION_BETWEEN=10

N_EXPERIMENTS=$((${#CLIENTS[@]} * (${#CLIENTS[@]} - 1) * $N_REPETITIONS * ${#N_INTERFACES[@]} * ${#PATH_SELECTION[@]}))
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
printf "Path selection:\t"; echo "${PATH_SELECTION[@]}"
printf "Duration:\t%s\n" "$DURATION"
printf "\nTotal number of experiments:\t%d\n" "$N_EXPERIMENTS"
printf "Estimated duration:\t\t%dh\n\n" "$TOTAL_DURATION"
########################################################################################################################
## Do a simple test to see if we are able to establish a connection
#  to all clients and clean up everything.
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/cleanUp.sh"
for CLIENT in "${CLIENTS[@]}"; do
 ssh -tt scion@"$CLIENT" -q "$SCRIPT_CMD" 1
 if [[ $? -ne 0 ]]; then
   printf "Failure : Cannot connect to %s.\n" "$CLIENT"
   exit 1
 fi
 printf "Success : Connection to %s.\n" "$CLIENT"
done
########################################################################################################################
## Initialize the clients
SCRIPT_NAME="init"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"

for CLIENT in "${CLIENTS[@]}"; do
 ssh -tt scion@"$CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
done
for CLIENT in "${CLIENTS[@]}"; do
 ./waitforReturn.sh "$CLIENT" "$SCRIPT_NAME" 0 30   # Polling w/ timeout after 30 seconds.
 if [[ $? -eq 1 ]]; then
   exit 1
 fi
 printf "Success : Initialization of %s.\n" "$CLIENT"
done
########################################################################################################################
## Do connection checks.
#  Start the connection test servers.
SCRIPT_NAME="connTestServer"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"

for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
done
sleep 3
for CLIENT in "${CLIENTS[@]}"; do
 ./waitforReturn.sh "$CLIENT" "$SCRIPT_NAME" 1 0   # No polling.
 if [[ $? -eq 1 ]]; then
   exit 1
 fi
 printf "Success : Starting connection test server on %s.\n" "$CLIENT"
done

#  Run the connection test clients.
SCRIPT_NAME="connTestClient"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"

for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
done
for CLIENT in "${CLIENTS[@]}"; do
 ./waitforReturn.sh "$CLIENT" "$SCRIPT_NAME" 0 20   # Polling w/ timeout 10 seconds.
 if [[ $? -eq 1 ]]; then
   exit 1
 fi
 printf "Success : Connection test for %s.\n" "$CLIENT"
done
########################################################################################################################