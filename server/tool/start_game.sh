#!/bin/sh


# 进入脚本所在目录
current_dir="$(dirname "$0")"
cd "$current_dir/.."

go build -o bin/game_server ./game/main.go

set -e
export RPC_LISTEN_ADDR=:5001
export RPC_ADDR=10.8.9.1:5001
export ETC_ENDPOINTS=127.0.0.1:2379
export MONGO_URL=mongodb://localhost:27017
./bin/game_server
