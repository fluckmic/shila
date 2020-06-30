#!/bin/bash

PRINT_DEBUG=1

PATH_TO_EXPERIMENT="~/go/src/shila/measurements/performance"

EXPERIMENT_NAME="Performance measurement"

mapfile -t CLIENTS < hostNames.data

CNT=0
for CLIENT in "${CLIENTS[@]}"; do
  CLIENT_IDS+=($CNT)
  CNT=$(($CNT+1))
done

N_REPETITIONS=10
N_INTERFACES=(1 2 3 5 7 10)
PATH_SELECTIONS=(0)
DURATION=210

DURATION_BETWEEN=20

N_EXPERIMENTS=$((${#CLIENTS[@]} * (${#CLIENTS[@]} - 1) * $N_REPETITIONS * ${#N_INTERFACES[@]} * ${#PATH_SELECTIONS[@]}))
TOTAL_DURATION_M=$(((($DURATION + $DURATION_BETWEEN) * $N_EXPERIMENTS) / 60 ))
TOTAL_DURATION_H=$(((($DURATION + $DURATION_BETWEEN) * $N_EXPERIMENTS) / 3600 ))

START_SESSION="bash ~/go/src/shila/measurements/sessionScripts/startSession.sh"
CHECK_ERROR="bash ~/go/src/shila/measurements/sessionScripts/checkForError.sh"

clear
rm -f -d -r _*
########################################################################################################################
## Print infos about the experiment.

printf "Starting %s:\n\n" "$EXPERIMENT_NAME"
printf "Clients:\t"; echo "${CLIENTS[@]}"
printf "Interfaces:\t"; echo "${N_INTERFACES[@]}"
printf "Path selection:\t"; echo "${PATH_SELECTIONS[@]}"
printf "Duration:\t%s\n" "$DURATION"
printf "\nTotal number of experiments:\t%d\n" "$N_EXPERIMENTS"
printf "Estimated duration:\t\t%dmin (%d h)\n\n" "$TOTAL_DURATION_M" "$TOTAL_DURATION_H"
########################################################################################################################
## Create the output folder / path.

DATE=$(date +%F-%H-%M-%S)
OUTPUT_PATH="_""$DATE"
mkdir "$OUTPUT_PATH"

########################################################################################################################
## Create the experiments file

./printDebug.sh "Creating the experiments file." "$PRINT_DEBUG"

rm -f _experiments.data

for SRC_ID in "${CLIENT_IDS[@]}"; do
  for DST_ID in "${CLIENT_IDS[@]}"; do
  if [[ $SRC_ID != $DST_ID ]]; then
    for N_INTERFACE in "${N_INTERFACES[@]}"; do
      for PATH_SELECT in "${PATH_SELECTIONS[@]}"; do
        COUNT=1
        while [[ "$COUNT" -le "$N_REPETITIONS" ]]; do
          echo "$SRC_ID" "$DST_ID" "$N_INTERFACE" "$PATH_SELECT" "$COUNT" >> _experiments.data
          COUNT=$(($COUNT+1))
        done
      done
    done
  fi
  done
done

# Create a random order
shuf _experiments.data | shuf -o _experiments.data

########################################################################################################################
## Initialize the clients
SCRIPT_NAME="init"
SCRIPT_CMD="sudo bash ""$PATH_TO_EXPERIMENT""/""$SCRIPT_NAME"".sh"

for CLIENT in "${CLIENTS[@]}"; do
 ./printDebug.sh "Start initializing ""$CLIENT""." "$PRINT_DEBUG"
 ssh -tt scion@"$CLIENT" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
 if [[ $? -ne 0 ]]; then
  printf "Failure : Cannot connect to %s.\n" "$CLIENT"
  exit 1
 fi
done
for CLIENT in "${CLIENTS[@]}"; do
 ./waitForReturn.sh "$CLIENT" "$SCRIPT_NAME" 0 30   # Polling w/ timeout after 30 seconds.
 if [[ $? -eq 1 ]]; then
   exit 1
 fi
 printf "Success : Initialization of %s.\n" "$CLIENT"
done
########################################################################################################################
## Run the experiment

rm _experiments.fail 2>/dev/null

N_EXPERIMENTS_DONE=0
N_EXPERIMENTS_FAIL=0

./printDebug.sh "Start doing experiments." "$PRINT_DEBUG"

while [[ "$N_EXPERIMENTS_DONE" != "$N_EXPERIMENTS" ]]; do

  # Repeat until all experiments are finished. Repeat the ones failed.
  if [[ $N_EXPERIMENTS_FAIL -gt 0 ]]; then

    rm _experiments.data
    cp _experiments.fail _experiments.data
    rm _experiments.fail
  fi
  N_EXPERIMENTS_FAIL=0

  EXPERIMENTS=()
  mapfile -t EXPERIMENTS < _experiments.data

  for EXPERIMENT in "${EXPERIMENTS[@]}"; do

    printf "Start with experiment %s.\n" "$EXPERIMENT"

    bash doExperiment.sh $EXPERIMENT "$DURATION" "$OUTPUT_PATH"
    if [[ $? -ne 0 ]]; then
      echo "$EXPERIMENT" >> _experiments.fail
      N_EXPERIMENTS_FAIL=$(($N_EXPERIMENTS_FAIL+1))
      printf "Failure : Experiment %s failed.\n" "$EXPERIMENT"
    else
      N_EXPERIMENTS_DONE=$(($N_EXPERIMENTS_DONE+1))
      printf "Success : Completed %d of %d experiments.\n" "$N_EXPERIMENTS_DONE" "$N_EXPERIMENTS"
    fi

  done
done
########################################################################################################################

















