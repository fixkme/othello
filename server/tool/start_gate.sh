#!/bin/sh


# 进入脚本所在目录
current_dir="$(dirname "$0")"
cd "$current_dir/.."

go build -o bin/gate_server ./gate/main.go

set -e
export GATE_LISTEN_ADDR=:7070
export RPC_LISTEN_ADDR=:5000
export RPC_ADDR=10.8.9.1:5000
./bin/gate_server
