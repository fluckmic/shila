#!/bin/bash

CHECK_SESSION="bash ~/go/src/shila/measurements/sessionScripts/isRunningSession.sh"
CHECK_ERROR="bash ~/go/src/shila/measurements/sessionScripts/checkForError.sh"

CLIENT=$1
SESSION_NAME=$2
POLL=$3         # Does poll the the session until it terminates if == 0
TIMEOUT=$4      # Just if polling is active, applies if > 0
STEP=0
if [[ "$TIMEOUT" -gt 0 ]]; then
  STEP=1
fi

# Polls the session until it terminates.
COUNT=0
while [ "$POLL" -eq 0  ]; do
  ssh -tt scion@"$CLIENT" -q "$CHECK_SESSION" "$SESSION_NAME"
  POLL=$?
  sleep 1
  COUNT="$COUNT"+"$STEP"
    if [[ "$COUNT" -gt "$TIMEOUT" ]]; then
      printf "Failure : Time out running %s for %s.\n" "$SESSION_NAME" "$CLIENT"
      exit 1
    fi
done

# Checks if the session terminated with an error. Returns error code 1 if so.
ssh -tt scion@"$CLIENT" -q "$CHECK_ERROR" "$SESSION_NAME" "$PATH_TO_EXPERIMENT"
if [[ $? -ne 0 ]]; then
  exit 1
fi

exit 0