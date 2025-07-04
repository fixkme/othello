#!/bin/bash

# 进入脚本所在目录
current_dir="$(dirname "$0")"
cd "$current_dir"

# 生成目录
OUT_DIR="../assets/main/scripts/pb"
PROTO_DIR="../../common/proto"

# 清理旧文件
rm -rf $OUT_DIR/*

# 生成 TS 代码
protoc \
  --plugin=protoc-gen-ts_proto=../node_modules/.bin/protoc-gen-ts_proto \
  --ts_proto_out=$OUT_DIR \
  --ts_proto_opt=outputEncodeMethods=true,outputJsonMethods=true,outputTypeRegistry=true,env=browser \
  -I=$PROTO_DIR \
  $PROTO_DIR/ws/*.proto $PROTO_DIR/datas/*.proto $PROTO_DIR/game/*.proto

# 统一文件权限（解决跨平台问题）
chmod -R 755 $OUT_DIR