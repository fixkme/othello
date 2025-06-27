#!/bin/sh

SERVICE_NAME=$1
MODE=$2

echo `pwd`

if [ ! -n "$SERVICE_NAME" ]; then
    echo "SERVICE_NAME is empty"
    exit 1
fi

if [ "$MODE" = "dlv" ]; then
  echo "Starting $SERVICE_NAME in dlv mode ..."
  DLV_PORT=${3:-2345}
  exec dlv debug --headless --listen=:${DLV_PORT} --api-version=2 --accept-multiclient "./$SERVICE_NAME"
else
  echo "Starting $SERVICE_NAME in normal mode $3"
  #go run "./$SERVICE_NAME/main.go"
  FLAG=$3
  go build ${FLAG} -o "./bin/$SERVICE_NAME" "./$SERVICE_NAME/main.go"
  exec "./bin/$SERVICE_NAME" >> "/app/logs/${SERVICE_NAME}_stdout.log" 2>> "/app/logs/${SERVICE_NAME}_stderr.log"
fi
