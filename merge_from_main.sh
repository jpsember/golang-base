#!/usr/bin/env bash
set -eu

echo "Attempting to merge from main into our branch (no-db)"
echo
echo "See: https://stackoverflow.com/questions/15232000/git-ignore-files-during-merge"

a=`git branch --show-current`
echo $a

if [ "$a" != "no-db" ]; then
  echo "We are on branch '$a', not 'no-db'!!!!"
  exit 1
fi

echo "More to come"
# git merge --no-ff --no-commit <merge-branch>
# git reset HEAD myfile.txt
# git checkout -- myfile.txt
# git commit -m "merged <merge-branch>"
