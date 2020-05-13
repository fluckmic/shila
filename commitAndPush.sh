#!/bin/sh

clear

if [ "$#" -ne 1 ]; then
  printf "Please provide commit message. \n"
  exit
fi

git fetch
git add --all
git commit -m "$1"
git push