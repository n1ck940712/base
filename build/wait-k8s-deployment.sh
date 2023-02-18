#!/bin/bash

WAIT_COUNT=0
MAX_WAIT_COUNT=900
DEPLOYMENT=
K8S_NAMESPACE=

while getopts "d:n:" option; do
  case $option in
    d )
      DEPLOYMENT=$OPTARG
      ;;
    n )
      K8S_NAMESPACE=$OPTARG
      ;;
  esac
done


while [ $(kubectl get -n $K8S_NAMESPACE deployment/$DEPLOYMENT -o 'jsonpath={$.status.replicas}') != $(kubectl get -n $K8S_NAMESPACE deployment/$DEPLOYMENT -o 'jsonpath={$.status.updatedReplicas}') ]
do
    echo 'Waiting for all replica to be updated'
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT+1))
    if [ "$WAIT_COUNT" -ge "$MAX_WAIT_COUNT" ]
    then
        echo "Timeout waiting for all replica to be updated"
        exit 1
    fi
done

WAIT_COUNT=0
while [ $(kubectl get -n $K8S_NAMESPACE deployment/$DEPLOYMENT -o 'jsonpath={$.status.replicas}') != $(kubectl get -n $K8S_NAMESPACE deployment/$DEPLOYMENT -o 'jsonpath={$.status.readyReplicas}') ]
do
    echo 'Waiting for all replica to be ready'
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT+1))
    if [ "$WAIT_COUNT" -ge "$MAX_WAIT_COUNT" ]
    then
        echo "Timeout waiting for all replica to be ready"
        exit 1
    fi
done
