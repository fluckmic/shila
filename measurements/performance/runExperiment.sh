#!/bin/bash

PRINT_DEBUG=1

PATH_TO_EXPERIMENT="~/go/src/shila/measurements/performance"
START_SESSION="bash ~/go/src/shila/measurements/sessionScripts/startSession.sh"

SRC_ID=$1; DST_ID=$2; DIRECTION=$3; N_INTERFACE=$4; PATH_SELECT=$5; REPETITION=$6
VALUE=$7; OUTPUT_FOLDER=$8; LOGFILE_EXPERIMENT=$9; MODE=$10; TIMEOUT=$11

mapfile -t CLIENTS < hostNames.data

SRC_CLIENT="${CLIENTS["$SRC_ID"]}"
DST_CLIENT="${CLIENTS["$DST_ID"]}"

LOG_FOLDER="$SRC_ID""_""$DST_ID""_""$DIRECTION""_""$N_INTERFACE""_""$PATH_SELECT""_""$REPETITION"
LOG_FILE="_iperfClientSide_""$LOG_FOLDER"".log"

########################################################################################################################
## Clean up the clients involved.
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/cleanUp.sh 0"
ssh -tt scion@"$SRC_CLIENT" -q "$SCRIPT_CMD"
if [[ $? -ne 0 ]]; then
    printf "Failure : Cannot connect to %s.\n" "$SRC_CLIENT" | tee -a "$LOGFILE_EXPERIMENT"
    exit 1
fi
ssh -tt scion@"$DST_CLIENT" -q "$SCRIPT_CMD"
if [[ $? -ne 0 ]]; then
    printf "Failure : Cannot connect to %s.\n" "$DST_CLIENT" | tee -a "$LOGFILE_EXPERIMENT"
    exit 1
fi
sleep 2
./printDebug.sh "Cleaned up the involved clients." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
sleep 10
########################################################################################################################
## Start the shila instance on the server.
./printDebug.sh "Starting shila instance on the server." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
SCRIPT_NAME="shilaServerSide"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
ssh -tt scion@"$DST_CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$DST_CLIENT" | tee -a "$LOGFILE_EXPERIMENT"
  exit 1
fi
sleep 6
./waitForReturn.sh "$DST_CLIENT" "$SCRIPT_NAME" 1 0   # No polling..
if [[ $? -eq 1 ]]; then
  exit 1
fi
sleep 2
./printDebug.sh "Done." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
########################################################################################################################
## Start the iperf instance on the server.
./printDebug.sh "Starting iperf instance on the server." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
SCRIPT_NAME="iperfServerSide"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
ssh -tt scion@"$DST_CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$DST_CLIENT" | tee -a "$LOGFILE_EXPERIMENT"
  exit 1
fi
sleep 5
./waitForReturn.sh "$DST_CLIENT" "$SCRIPT_NAME" 1 0   # No polling..
if [[ $? -eq 1 ]]; then
  exit 1
fi
sleep 2
./printDebug.sh "Done." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
########################################################################################################################
## Start the shila instance on the client.
./printDebug.sh "Starting shila instance on the client." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
SCRIPT_NAME="shilaClientSide"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
ssh -tt scion@"$SRC_CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD" "$N_INTERFACE" "$PATH_SELECT"
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$SRC_CLIENT" | tee -a "$LOGFILE_EXPERIMENT"
  exit 1
fi
sleep 6
./waitForReturn.sh "$SRC_CLIENT" "$SCRIPT_NAME" 1 0   # No polling..
if [[ $? -eq 1 ]]; then
  exit 1
fi
./printDebug.sh "Done." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
sleep 10
########################################################################################################################
## Start the iperf instance on the client.
./printDebug.sh "Starting iperf instance on the client." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
SCRIPT_NAME="iperfClientSide"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"
ssh -tt scion@"$SRC_CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD" "$DST_ID" "$DIRECTION" "$N_INTERFACE" "$PATH_SELECT" "$REPETITION" "$VALUE" $MODE
if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$SRC_CLIENT" | tee -a "$LOGFILE_EXPERIMENT"
  exit 1
fi
./waitForReturn.sh "$SRC_CLIENT" "$SCRIPT_NAME" 0 "$TIMEOUT"   # With polling, times out after certain time.
if [[ $? -eq 1 ]]; then
  ERROR_DATE=$(date +%F-%H-%M-%S)
  ERROR_FOLDER="$OUTPUT_FOLDER""/""$LOG_FOLDER""_""$ERROR_DATE"

  mkdir "$ERROR_FOLDER"

  scp scion@"$SRC_CLIENT":"$PATH_TO_EXPERIMENT""/_*.log" "$ERROR_FOLDER"
  scp scion@"$SRC_CLIENT":"$PATH_TO_EXPERIMENT""/_*.err" "$ERROR_FOLDER"
  scp scion@"$SRC_CLIENT":"$PATH_TO_EXPERIMENT""/_*.dump" "$ERROR_FOLDER"
  scp scion@"$DST_CLIENT":"$PATH_TO_EXPERIMENT""/_*.log" "$ERROR_FOLDER"
  scp scion@"$DST_CLIENT":"$PATH_TO_EXPERIMENT""/_*.err" "$ERROR_FOLDER"
  scp scion@"$DST_CLIENT":"$PATH_TO_EXPERIMENT""/_*.dump" "$ERROR_FOLDER"

  exit 1
fi
./printDebug.sh "Done." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
########################################################################################################################
## Copy back the measurements

mkdir "$OUTPUT_FOLDER""/""$LOG_FOLDER"

./printDebug.sh "Start copying back the experiment data." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"

scp scion@"$SRC_CLIENT":"$PATH_TO_EXPERIMENT"/"$LOG_FILE" "$OUTPUT_FOLDER""/""$LOG_FOLDER"
if [[ $? -ne 0 ]]; then
  printf "Failure : Unable to copy back the result from %s.\n" "$SRC_CLIENT" | tee -a "$LOGFILE_EXPERIMENT"
  exit 1
fi

scp scion@"$SRC_CLIENT":"$PATH_TO_EXPERIMENT""/_shilaClientSide.log" "$OUTPUT_FOLDER""/""$LOG_FOLDER"
scp scion@"$SRC_CLIENT":"$PATH_TO_EXPERIMENT""/_clientConfig.dump"   "$OUTPUT_FOLDER""/""$LOG_FOLDER"
scp scion@"$DST_CLIENT":"$PATH_TO_EXPERIMENT""/_shilaServerSide.log" "$OUTPUT_FOLDER""/""$LOG_FOLDER"
scp scion@"$DST_CLIENT":"$PATH_TO_EXPERIMENT""/_serverConfig.dump"   "$OUTPUT_FOLDER""/""$LOG_FOLDER"

./printDebug.sh "Done." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"