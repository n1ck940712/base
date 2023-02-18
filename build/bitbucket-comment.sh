#!/bin/bash

STATUS=$1
BITBUCKET_USERNAME=$2
BITBUCKET_PASSWORD=$3
BRANCH_NAME=$4
GIT_URL=$5

REPO_SLUG=`echo $GIT_URL | sed -n -e 's/.*esportsph\/\(.*\)\.git/\1/p'`

echo "Checking PR for $REPO_SLUG with source branch $BRANCH_NAME"

PR_NUMBER=`curl --request GET \
  --url 'https://api.bitbucket.org/2.0/repositories/esportsph/'$REPO_SLUG'/pullrequests?q=state="OPEN"+AND+source.branch.name="'$BRANCH_NAME'"' \
  -u "$BITBUCKET_USERNAME:$BITBUCKET_PASSWORD" \
  --header 'Accept: application/json' |  jq '.values[0].id'`
if [ "$PR_NUMBER" == "null" ]; then
    echo "No PR found"
    exit 0
fi

if [ "$STATUS" == "success" ]; then
    curl --request POST \
      --url 'https://api.bitbucket.org/2.0/repositories/esportsph/'$REPO_SLUG'/pullrequests/'$PR_NUMBER'/comments/' \
      -u "$BITBUCKET_USERNAME:$BITBUCKET_PASSWORD" \
      --header 'Accept: application/json' \
      --header 'Content-Type: application/json' \
      -d '{"content": {"raw": "Jenkins build: 🟢 🚀 🚀 Build success"}}'
elif [ "$STATUS" == "failed" ]; then
    curl --request POST \
      --url 'https://api.bitbucket.org/2.0/repositories/esportsph/'$REPO_SLUG'/pullrequests/'$PR_NUMBER'/comments/' \
      -u "$BITBUCKET_USERNAME:$BITBUCKET_PASSWORD" \
      --header 'Accept: application/json' \
      --header 'Content-Type: application/json' \
      -d '{"content": {"raw": "Jenkins build: 🔴 ✋ ✋ Build failed"}}'
else
    echo "Unsupported status"
fi
