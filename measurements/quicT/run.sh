#!/bin/bash

PRINT_DEBUG=1

DRY_RUN=$1

PATH_TO_EXPERIMENT="~/go/src/shila/measurements/quicT"

EXPERIMENT_NAME="performance measurment quic"

mapfile -t CLIENTS < hostNames.data

CLIENT_IDS=(0 1 3)
N_REPETITIONS=10
DIRECTIONS=(0 1)    # 0: client -> server
                    # 1: server -> client

TRANSFER=100             # Amount of data (in MByte) to send

DURATION_BETWEEN=60      # For estimating the duration of the experiment (Seconds).
DURATION=$(($TRANSFER))  # For estimating the duration of the experiment (MByte/s). Assume 1 MByte/s throughput..

# Factor two because of bi-direction tests client -> server, and client <- server
N_EXPERIMENTS=$((${#CLIENT_IDS[@]} * (${#CLIENT_IDS[@]} - 1) * $N_REPETITIONS * 2))
TOTAL_DURATION_M=$(((($DURATION + $DURATION_BETWEEN) * $N_EXPERIMENTS) / 60 ))
TOTAL_DURATION_H=$(((($DURATION + $DURATION_BETWEEN) * $N_EXPERIMENTS) / 3600 ))

START_SESSION="bash ~/go/src/shila/measurements/sessionScripts/startSession.sh"
CHECK_ERROR="bash ~/go/src/shila/measurements/sessionScripts/checkForError.sh"

clear
rm -f -d -r _*
########################################################################################################################
## Create the output folder / path.

DATE=$(date +%F-%H-%M-%S)
OUTPUT_PATH="_""$DATE"
mkdir "$OUTPUT_PATH"
########################################################################################################################
## Print infos about the experiment.

LOGFILE_EXPERIMENT="$OUTPUT_PATH""/experiment.log"

printf "Starting %s:\n\n" "$EXPERIMENT_NAME" | tee -a "$LOGFILE_EXPERIMENT"

printf "Clients:\t" | tee -a "$LOGFILE_EXPERIMENT"
for CLIENT_ID in "${CLIENT_IDS[@]}"; do
printf "%s " "${CLIENTS[$CLIENT_ID]}" | tee -a "$LOGFILE_EXPERIMENT"
done
printf "\n" | tee -a "$LOGFILE_EXPERIMENT"

printf "Client Ids:\t" | tee -a "$LOGFILE_EXPERIMENT"
for CLIENT_ID in "${CLIENT_IDS[@]}"; do
printf "%s " "$CLIENT_ID" | tee -a "$LOGFILE_EXPERIMENT"
done
printf "\n" | tee -a "$LOGFILE_EXPERIMENT"

printf "Repetitions:\t%s\n" "$N_REPETITIONS" | tee -a "$LOGFILE_EXPERIMENT"
printf "Duration:\t%s\n" "$DURATION" | tee -a "$LOGFILE_EXPERIMENT"
printf "\nTotal number of experiments:\t%d\n" "$N_EXPERIMENTS" | tee -a "$LOGFILE_EXPERIMENT"
printf "Estimated duration:\t\t%dmin (%d h)\n\n" "$TOTAL_DURATION_M" "$TOTAL_DURATION_H" | tee -a "$LOGFILE_EXPERIMENT"
########################################################################################################################
## Create the experiments file
./printDebug.sh "Creating the experiments file." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"

rm -f _experiments_0.data

for SRC_ID in "${CLIENT_IDS[@]}"; do
  for DST_ID in "${CLIENT_IDS[@]}"; do
  if [[ $SRC_ID != $DST_ID ]]; then
    for DIRECTION in "${DIRECTIONS[@]}"; do
          COUNT=1
          while [[ "$COUNT" -le "$N_REPETITIONS" ]]; do
            echo "$SRC_ID" "$DST_ID" "$DIRECTION" "$COUNT" >> _experiments_0.data
            COUNT=$(($COUNT+1))
      done
    done
  fi
  done
done

# Create a random order
shuf _experiments_0.data | shuf -o _experiments_0.data

if [[ $DRY_RUN -eq 1 ]]; then
  exit 0
fi

########################################################################################################################
## Initialize the clients
SCRIPT_NAME="init"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"

for CLIENT_ID in "${CLIENT_IDS[@]}"; do
 ./printDebug.sh "Start initializing ""${CLIENTS[$CLIENT_ID]}""." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"
 ssh -tt scion@"${CLIENTS[$CLIENT_ID]}" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD" "$CONGESTION_CONTROL"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "${CLIENTS[$CLIENT_ID]}" | tee -a "$LOGFILE_EXPERIMENT"
  exit 1
 fi
done
for CLIENT_ID in "${CLIENT_IDS[@]}"; do
 ./waitForReturn.sh "${CLIENTS[$CLIENT_ID]}" "$SCRIPT_NAME" 0 30   # Polling w/ timeout after 30 seconds.
 if [[ $? -eq 1 ]]; then
   exit 1
 fi
 printf "Success : Initialization of %s.\n" "${CLIENTS[$CLIENT_ID]}" | tee -a "$LOGFILE_EXPERIMENT"
done

########################################################################################################################
## Run the experiment

rm _*.fail 2>/dev/null

N_EXPERIMENTS_DONE=0
N_EXPERIMENTS_FAIL=0
N_REPS=0

./printDebug.sh "Start doing experiments." "$PRINT_DEBUG" "$LOGFILE_EXPERIMENT"

while [[ $N_EXPERIMENTS_DONE -lt $N_EXPERIMENTS ]]; do

  # Repeat until all experiments are finished. Repeat the ones failed.
  if [[ $N_EXPERIMENTS_FAIL -gt 0 ]]; then
    cp "$FAIL_LOG" "_experiments_""$N_REPS".data
  fi
  N_EXPERIMENTS_FAIL=0
  FAIL_LOG="_experiments_""$N_REPS"".fail"

  EXPERIMENTS=()
  mapfile -t EXPERIMENTS < "_experiments_""$N_REPS"".data"

  for EXPERIMENT in "${EXPERIMENTS[@]}"; do

    printf "Start with experiment %s.\n" "$EXPERIMENT" | tee -a "$LOGFILE_EXPERIMENT"

    TIMEOUT=$(($DURATION * 2))
    bash runExperiment.sh $EXPERIMENT "$TRANSFER" "$OUTPUT_PATH" "$LOGFILE_EXPERIMENT" "$TIMEOUT"
    if [[ $? -ne 0 ]]; then
      echo "$EXPERIMENT" >> "$FAIL_LOG"
      N_EXPERIMENTS_FAIL=$(($N_EXPERIMENTS_FAIL+1))
      printf "Failure : Experiment %s failed.\n" "$EXPERIMENT" | tee -a "$LOGFILE_EXPERIMENT"
    else
      N_EXPERIMENTS_DONE=$(($N_EXPERIMENTS_DONE+1))
      printf "Success : Completed %d of %d experiments.\n" "$N_EXPERIMENTS_DONE" "$N_EXPERIMENTS" | tee -a "$LOGFILE_EXPERIMENT"
    fi
  done

  N_REPS=$(($N_REPS+1))

done
########################################################################################################################
