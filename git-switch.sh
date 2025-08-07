#!/bin/sh
if [ "$1" = "work" ]; then
  git config --global user.name "Prasan Internshala"
  git config --global user.email "prasan@internshala.com"
  echo "Switched to WORK account"
elif [ "$1" = "personal" ]; then
  git config --global user.name "Prasan"
  git config --global user.email "prasanmishra330@gmail.com"
  echo "Switched to PERSONAL account"
else
  echo "Usage: ./git-switch.sh [work|personal]"
fi
exit
