#!/bin/bash

# sudo sysctl net.mptcp.mptcp_scheduler=default
# sudo sysctl net.ipv4.tcp_congestion_control=lia
# sudo sysctl net.ipv4.tcp_congestion_control=cubic

PRINT_DEBUG=1

DRY_RUN=$1

PATH_TO_EXPERIMENT="~/go/src/shila/measurements/performance"

EXPERIMENT_NAME="performance measurement"

mapfile -t CLIENTS < hostNames.data

CLIENT_IDS=(0 1 )
N_REPETITIONS=2
N_INTERFACES=(2)
PATH_SELECTIONS=(1)
DIRECTIONS=(0 1)    # 0: client -> server
                    # 1: server -> client

DURATION=25   # How long to send data
TRANSFER=10   # Amount of data (in MByte) to send

DURATION_MODE=1
TRANSFER_MODE=2

MODE=$DURATION_MODE

if [[ $MODE -eq $DURATION_MODE ]]; then
  MODE_DESC="duration"
fi
if [[ $MODE -eq $TRANSFER_MODE ]]; then
  MODE_DESC="transfer"
fi

DURATION_BETWEEN=60   # For estimating the duration of the experiment (Seconds).
APPROX_THROUGHPUT=1   # For estimating the duration of the experiment (MByte/s).

if [[ $MODE -eq $TRANSFER_MODE ]]; then
  DURATION=$(($TRANSFER / $APPROX_THROUGHPUT))
fi

# Factor two because of bi-direction tests client -> server, and client <- server
N_EXPERIMENTS=$((${#CLIENT_IDS[@]} * (${#CLIENT_IDS[@]} - 1) * $N_REPETITIONS * ${#N_INTERFACES[@]} * ${#PATH_SELECTIONS[@]} * 2))
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

printf "Starting %s:\n\n" "$EXPERIMENT_NAME"" in ""$MODE_DESC"" mode" | tee -a "$LOGFILE_EXPERIMENT"

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

printf "Interfaces:\t" | tee -a "$LOGFILE_EXPERIMENT"
for N_INTERFACE in "${N_INTERFACES[@]}"; do
printf "%s " "$N_INTERFACE" | tee -a "$LOGFILE_EXPERIMENT"
done
printf "\n" | tee -a "$LOGFILE_EXPERIMENT"

printf "Path selection:\t" | tee -a "$LOGFILE_EXPERIMENT"
for PATH_SELECT in "${PATH_SELECTIONS[@]}"; do
printf "%s " "$PATH_SELECT" | tee -a "$LOGFILE_EXPERIMENT"
done
printf "\n" | tee -a "$LOGFILE_EXPERIMENT"

printf "Repetitions:\t%s\n" "$N_REPETITIONS" | tee -a "$LOGFILE_EXPERIMENT"

if [[ $MODE -eq $DURATION_MODE ]]; then
  printf "Duration:\t%s\n" "$DURATION" | tee -a "$LOGFILE_EXPERIMENT"
fi

if [[ $MODE -eq $TRANSFER_MODE ]]; then
  printf "Transfer:\t%s\n" "$TRANSFER" | tee -a "$LOGFILE_EXPERIMENT"
fi

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
      for N_INTERFACE in "${N_INTERFACES[@]}"; do
        for PATH_SELECT in "${PATH_SELECTIONS[@]}"; do
          COUNT=1
          while [[ "$COUNT" -le "$N_REPETITIONS" ]]; do
            echo "$SRC_ID" "$DST_ID" "$DIRECTION" "$N_INTERFACE" "$PATH_SELECT" "$COUNT" >> _experiments_0.data
            COUNT=$(($COUNT+1))
          done
        done
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
 ssh -tt scion@"${CLIENTS[$CLIENT_ID]}" -q "$START_SESSION" "$SCRIPT_NAME" "$SCRIPT_CMD"
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

    if [[ $MODE -eq $DURATION_MODE ]]; then
      VALUE=$DURATION
      TIMEOUT=$(($DURATION * 2))
    fi
    if [[ $MODE -eq $TRANSFER_MODE ]]; then
      VALUE=$TRANSFER
      TIMEOUT=$(($DURATION * 3))
    fi
    bash runExperiment.sh $EXPERIMENT "$VALUE" "$OUTPUT_PATH" "$LOGFILE_EXPERIMENT" "$MODE"
    if [[ $? -ne 0 ]]; then
      echo "$EXPERIMENT" >> "$FAIL_LOG"
      N_EXPERIMENTS_FAIL=$(($N_EXPERIMENTS_FAIL+1))
      printf "Failure : Experiment %s failed.\n" "$EXPERIMENT" | tee -a "$LOGFILE_EXPERIMENT"
    else
      N_EXPERIMENTS_DONE=$(($N_EXPERIMENTS_DONE+1))
      printf "Success : Completed %d of %d experiments.\n" "$N_EXPERIMENTS_DONE" "$N_EXPERIMENTS" | tee -a "$LOGFILE_EXPERIMENT"
    fi
  done

  exit 0

  N_REPS=$(($N_REPS+1))

done
########################################################################################################################
