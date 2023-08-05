#!/usr/bin/env bash
set -eu

EXPECTED_BRANCH=no-db
OTHER_BRANCH=main
CURRENT_BRANCH=`git branch --show-current`

echo "Current branch: $CURRENT_BRANCH"
echo "Attempting to merge from $OTHER_BRANCH into $EXPECTED_BRANCH"
echo
echo "See: https://stackoverflow.com/questions/15232000/git-ignore-files-during-merge"
echo


if [ "$CURRENT_BRANCH" != "$EXPECTED_BRANCH" ]; then
  echo "Current branch is $CURRENT_BRANCH, expected $EXPECTED_BRANCH !!!"
  exit 1
fi

OMITTED_FILE="webapp/database.go"
echo "Omitting: $OMITTED_FILE"

git merge --no-ff --no-commit main
git reset HEAD $OMITTED_FILE
git checkout -- $OMITTED_FILE
git commit -m "merged $OTHER_BRANCH"
