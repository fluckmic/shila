#!/bin/bash

DEST_IP="127.0.0.1"

clear

gcc client.c -o _client

./_client -c "$DEST_IP" -d
