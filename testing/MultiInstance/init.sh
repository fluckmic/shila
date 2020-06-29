#!/bin/bash

BASE_DIR=$(dirname "$0")
cd "$BASE_DIR"

# Kill running instances
pkill -f shila

# Delete all namespaces
bash ../../helper/netnsClear.sh

## Update the repo
git pull

