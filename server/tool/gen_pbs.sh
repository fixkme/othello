#!/bin/sh


# 进入脚本所在目录
current_dir="$(dirname "$0")"
cd "$current_dir"

# 生成目录
OUT_DIR="../pb"
PROTO_DIR="../../common/proto"
PROTO_SS_DIR="../../common/proto_ss"

# 清理旧文件
rm -rf $OUT_DIR/*

protoc -I $PROTO_DIR  --gom_out=$OUT_DIR \
--gom_opt=paths=source_relative,\
pbext-pkg=github.com/fixkme/othello/server/pb,\
data-pkgs=datas^models,\
rpc-pkgs=game \
${PROTO_DIR}/datas/*.proto ${PROTO_DIR}/models/*.proto ${PROTO_DIR}/game/*.proto 

