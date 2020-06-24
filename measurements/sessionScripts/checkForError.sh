#!/bin/bash

# Checks whether a certain session returned with an error.

SESSION_NAME=$1

ERROR_FILE=_"$SESSION_NAME"

if [[ -f "$ERROR_FILE" ]]; then
