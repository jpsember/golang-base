#!/usr/bin/env bash
set -eu

# Merge from 'other' branch to 'our' branch, leaving some files untouched.
#
# The two branches must be one of "main" or "no-db".
#
# The untouched files are at present the single file "webapp/database.go"
#

MAIN_BRANCH="main"
ALT_BRANCH="no-db"

CURRENT_BRANCH=`git branch --show-current`
if [ "$CURRENT_BRANCH" == "$MAIN_BRANCH" ]; then
  OTHER_BRANCH=$ALT_BRANCH
elif [ "$CURRENT_BRANCH" == "ALT_BRANCH" ]; then
  OTHER_BRANCH=$MAIN_BRANCH
else
  echo "Current branch is $CURRENT_BRANCH, expected either $MAIN_BRANCH or $ALT_BRANCH !!!"
  exit 1
fi

echo "Current branch: $CURRENT_BRANCH"
echo "Attempting to merge from $OTHER_BRANCH into $CURRENT_BRANCH"
echo
echo "See: https://stackoverflow.com/questions/15232000/git-ignore-files-during-merge"
echo

OMITTED_FILE="webapp/database.go"
echo "Omitting: $OMITTED_FILE"

echo "Exiting"
exit 1
git merge --no-ff --no-commit $OTHER_BRANCH
git reset HEAD $OMITTED_FILE
git checkout -- $OMITTED_FILE
git commit -m "merged $OTHER_BRANCH"
