#!/bin/sh


# 进入脚本所在目录
current_dir="$(dirname "$0")"
cd "$current_dir/.."

go build -o bin/hall_server ./hall/main.go

set -e
export SERVER_ID=1
export RPC_LISTEN_ADDR=:5001
export RPC_ADDR=localhost:5001
export MONGO_URI=mongodb://localhost:27017
./bin/hall_server
