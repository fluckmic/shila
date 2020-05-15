#!/bin/sh

clear

printf "Fetch newest version from gitlab..\n"
git pull

printf "Compile..\n"
go build

printf "Run shila..\n"
sudo ./shila

