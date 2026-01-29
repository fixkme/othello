#!/bin/sh


# 进入脚本所在目录
current_dir="$(dirname "$0")"
cd "$current_dir/.."

go build -o bin/game_server ./game/main.go

set -e
export SERVER_ID=1
export RPC_LISTEN_ADDR=:5002
export RPC_ADDR=10.8.9.1:5002
export MONGO_URI=mongodb://localhost:27017
./bin/game_server
