#!/bin/bash

if [ "$(uname)" == "Darwin" ]; then
    SED=gsed
else
    SED=sed
fi

# 进入脚本所在目录
current_dir="$(dirname "$0")"
cd "$current_dir"

# 给定文件夹路径
in_dir="../../common/proto/"
go_mod_name="github.com/fixkme/othello/server/pb"
in_paths=(
  "hall"
  "game"
)

output_file="./pb/register_message.go"

# 清空现有文件内容或创建新文件，如果存在则覆盖
> "$output_file"

echo -e 'package pb\n' >> "$output_file"
echo -e "import (" >> "$output_file"
for element in "${in_paths[@]}"
do
    im="\t_ \"${go_mod_name}/${element}\""
    echo -e $im >> "$output_file"
done
echo -e ")\n" >> "$output_file"

# 找出request和response、notice的名字
req_arr=()
rsp_arr=()
push_arr=()
for element in "${in_paths[@]}"; do
    folder_path="${in_dir}/${element}"
    if [ ! -d "$folder_path" ]; then
        echo "Directory does not exist: $folder_path" >&2
        continue
    fi

    # 使用进程替换 <(...) 避免子 shell
    while IFS= read -r -d $'\0' file; do
        # 读取文件中的 message 行
        while IFS= read -r line; do
            # 提取 message 后的单词
            extracted_word=$(echo "$line" | $SED -E 's/message[[:space:]]*([[:alnum:]_]+)/\1/i')
            
            # 跳过不符合条件的
            if [[ -z "$extracted_word" ]] || [[ ! $extracted_word =~ ^[A-Z]{2} ]]; then
                continue
            fi

            msg="${element}.${extracted_word}"

            if [[ $extracted_word == C* ]]; then
                req_arr+=("$msg")
            elif [[ $extracted_word == S* ]]; then
                rsp_arr+=("$msg")
            elif [[ $extracted_word == P* ]]; then
                push_arr+=("$msg")
            fi
        done < <(grep -i -o 'message[[:space:]]*[[:alnum:]_]*' "$file" | $SED -E 's/message[[:space:]]*([[:alnum:]_]+)/\1/i')
    done < <(find "$folder_path" -type f ! -name "*grpc*" -print0)
done

# 写入代码
echo -e "var RequestMsgNames = []string{" >> "$output_file"
for element in "${req_arr[@]}"
do
    echo -e "\t\"${element}\"," >> "$output_file"
done
echo -e "}\n" >> "$output_file"

echo -e "var ResponseMsgNames = []string{" >> "$output_file"
for element in "${rsp_arr[@]}"
do
    echo -e "\t\"${element}\"," >> "$output_file"
done
echo -e "}\n" >> "$output_file"

echo -e "var NoticeMsgNames = []string{" >> "$output_file"
for element in "${push_arr[@]}"
do
    echo -e "\t\"${element}\"," >> "$output_file"
done
echo -e "}\n" >> "$output_file"

echo "register protobuf message into $output_file"


# 编译
if [ "$1" == "build" ]; then
    go build -o "prototest" main.go
fi
