#!/bin/bash

CONNECTION_TEST_TIMEOUT=15

CLIENTS=(mptcp-over-scion-vm-1 mptcp-over-scion-vm-2)

START_SESSION='bash ~/go/src/shila/measurements/sessionScripts/startSession.sh'
CHECK_SESSION='bash ~/go/src/shila/measurements/sessionScripts/isRunningSession.sh'
CHECK_ERROR='bash ~/go/src/shila/measurements/sessionScripts/checkForError.sh'
KILL_ALL_SESSIONS='bash ~/go/src/shila/measurements/sessionScripts/terminateAllSessions.sh'

PATH_TO_EXPERIMENT="~/go/src/shila/measurements/performance"

clear

## Do a simple test to see if we are able to establish a connection to all
#  clients and kill all running tmux sessions.
for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q "$KILL_ALL_SESSIONS"
  if [[ $? -ne 0 ]]; then
    printf "Unable to connect to %s.\n" "$CLIENT"
    exit 1
  fi
  printf "Success : Connection to %s.\n" "$CLIENT"
done

## Initialize the clients
SCRIPT_NAME="init"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"

for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
done

for CLIENT in "${CLIENTS[@]}"; do
  RUNNING=0
  while [ "$RUNNING" -eq 0  ]; do
      ssh -tt scion@"$CLIENT" -q "$CHECK_SESSION" "$SCRIPT_NAME"
      RUNNING=$?
      sleep 1
  done

  ssh -tt scion@"$CLIENT" -q "$CHECK_ERROR" "$SCRIPT_NAME" "$PATH_TO_EXPERIMENT"
  if [[ $? -ne 0 ]]; then
    exit 1
  fi

  printf "Success : Initialization of %s.\n" "$CLIENT"
done

## Then do a connection checks.

#  Start the connection test servers.
SCRIPT_NAME="connTestServer"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"

for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
done

sleep 3

for CLIENT in "${CLIENTS[@]}"; do
  ssh -tt scion@"$CLIENT" -q "$CHECK_ERROR" "$SCRIPT_NAME" "$PATH_TO_EXPERIMENT"
  if [[ $? -ne 0 ]]; then
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
  RUNNING=0
  COUNT=0
  while [ "$RUNNING" -eq 0  ]; do
      ssh -tt scion@"$CLIENT" -q "$CHECK_SESSION" "$SCRIPT_NAME"
      RUNNING=$?
      sleep 1
      COUNT="$COUNT"+1
      if [[ "$COUNT" -ge "$CONNECTION_TEST_TIMEOUT" ]]; then
        printf "Error : Timeout in connection test for %s.\n" "$CLIENT"
        exit 1
      fi
  done

  ssh -tt scion@"$CLIENT" -q "$CHECK_ERROR" "$SCRIPT_NAME" "$PATH_TO_EXPERIMENT"
  if [[ $? -ne 0 ]]; then
    exit 1
  fi

  printf "Success : Connection test for %s.\n" "$CLIENT"
done
