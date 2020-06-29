#!/bin/bash

PRINT_DEBUG=1

PATH_TO_EXPERIMENT="~/go/src/shila/measurements/performance"
START_SESSION="bash ~/go/src/shila/measurements/sessionScripts/startSession.sh"

SRC_ID=$1; DST_ID=$2; N_INTERFACE=$3; PATH_SELECT=$4; REPETITION=$5; DURATION=$6; OUTPUT_FOLDER=$7

mapfile -t CLIENTS < hostNames.data

SRC_CLIENT="${CLIENTS["$SRC_ID"]}"
DST_CLIENT="${CLIENTS["$DST_ID"]}"

LOG_FOLDER="$SRC_ID""_""$DST_ID""_""$N_INTERFACE""_""$PATH_SELECT""_""$REPETITION"
LOG_FILE="_iperfClientSide_""$LOG_FOLDER"".log"

########################################################################################################################
## Clean up the clients involved.
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/cleanUp.sh 0"
ssh -tt scion@"$SRC_CLIENT" -q "$SCRIPT_CMD"
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$SRC_CLIENT"
  exit 1
fi
ssh -tt scion@"$DST_CLIENT" -q "$SCRIPT_CMD"
 if [[ $? -ne 0 ]]; then
   printf "Failure : Cannot connect to %s.\n" "$DST_CLIENT"
   exit 1
 fi
./printDebug.sh "Cleaned up the involved clients." "$PRINT_DEBUG"
########################################################################################################################
## Start the shila instance on the server.
SCRIPT_NAME="shilaServerSide"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
ssh -tt scion@"$DST_CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$DST_CLIENT"
  exit 1
fi
sleep 2
./waitForReturn.sh "$DST_CLIENT" "$SCRIPT_NAME" 1 0   # No polling..
if [[ $? -eq 1 ]]; then
  exit 1
fi
./printDebug.sh "Started shila instance on the server." "$PRINT_DEBUG"
########################################################################################################################
## Start the iperf instance on the server.
SCRIPT_NAME="iperfServerSide"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
ssh -tt scion@"$DST_CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$DST_CLIENT"
  exit 1
fi
sleep 2
./waitForReturn.sh "$DST_CLIENT" "$SCRIPT_NAME" 1 0   # No polling..
if [[ $? -eq 1 ]]; then
  exit 1
fi
./printDebug.sh "Started iperf instance on the server." "$PRINT_DEBUG"
########################################################################################################################
## Start the shila instance on the client.
SCRIPT_NAME="shilaClientSide"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
ssh -tt scion@"$SRC_CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD" "$N_INTERFACE" "$PATH_SELECT"
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$SRC_CLIENT"
  exit 1
fi
sleep 2
./waitForReturn.sh "$SRC_CLIENT" "$SCRIPT_NAME" 1 0   # No polling..
if [[ $? -eq 1 ]]; then
  exit 1
fi
./printDebug.sh "Started shila instance on the client." "$PRINT_DEBUG"
########################################################################################################################
## Start the iperf instance on the client.
SCRIPT_NAME="iperfClientSide"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
ssh -tt scion@"$SRC_CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD" "$DST_ID" "$N_INTERFACE" "$PATH_SELECT" "$REPETITION" "$DURATION"
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$SRC_CLIENT"
  exit 1
fi
./waitForReturn.sh "$SRC_CLIENT" "$SCRIPT_NAME" 0 0 #$((2 * $DURATION))   # With polling, times out after 2x experiment duration.
if [[ $? -eq 1 ]]; then
  exit 1
fi
./printDebug.sh "Iperf instance on the client completed successful." "$PRINT_DEBUG"
########################################################################################################################
## Copy back the measurements

mkdir "$OUTPUT_FOLDER""/""$LOG_FOLDER"

scp scion@"$SRC_CLIENT":"$PATH_TO_EXPERIMENT"/"$LOG_FILE" "$OUTPUT_FOLDER""/""$LOG_FOLDER"
if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to copy back the result from %s.\n" "$SRC_CLIENT"
  exit 1
fi

scp scion@"$SRC_CLIENT":"$PATH_TO_EXPERIMENT""/_shilaClientSide.log" "$OUTPUT_FOLDER""/""$LOG_FOLDER"
scp scion@"$DST_CLIENT":"$PATH_TO_EXPERIMENT""/_shilaServerSide.log" "$OUTPUT_FOLDER""/""$LOG_FOLDER"

if [[ -f _latestExperiment.log ]]; then
  rm _latestExperiment.log
fi

cp "$OUTPUT_FOLDER""/""$LOG_FOLDER""/""$LOG_FILE" _latestExperiment.log

./printDebug.sh "Copied back the experiment data." "$PRINT_DEBUG"
########################################################################################################################
## Clean up the clients involved.
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/cleanUp.sh"
ssh -tt scion@"$SRC_CLIENT" -q "$SCRIPT_CMD" 0
ssh -tt scion@"$DST_CLIENT" -q "$SCRIPT_CMD" 0

./printDebug.sh "Cleaned up the involved clients." "$PRINT_DEBUG"
########################################################################################################################