package logic

var ScriptTemplate = `
#! /bin/bash
#
# Copyright (c) 2009-2015
# All Rights Reserved
set -e
ARCHIVE=$(awk '/^__ARCHIVE_BELOW__/ {print NR + 1; exit 0; }' "$0")
tmp_dir=/tmp/extract_dir # 定义一个临时目录变量
rm -rf $tmp_dir
mkdir -p $tmp_dir
isencrypt={{.Isencrypt}}
signature={{.Signature}}
# 检查openssl是否存在
if ! which openssl >/dev/null; then
    echo "Error: openssl not found." >&2
    rm -rf $tmp_dir
    exit 1
fi

# 检查xxd是否存在
if ! which xxd >/dev/null; then
    echo "Error: xxd not found." >&2
    rm -rf $tmp_dir
    exit 1
fi

# 检查签名值是否不为空
if [ -n "$signature" ]; then
    # 创建签名文件
    echo "$signature" >${tmp_dir}/signature
    # 创建负载文件
    tail -n+$ARCHIVE "$0" >${tmp_dir}/payload
    xxd -r -p ${tmp_dir}/signature ${tmp_dir}/signature.bin
    # 使用openssl验证签名
    openssl dgst -verify $1 -signature ${tmp_dir}/signature.bin ${tmp_dir}/payload
    if [ $? -eq 0 ]; then
        echo "Signature verification succeeded."
    else
        echo "Error: Signature verification failed." >&2
        rm -rf $tmp_dir
        exit 1
    fi
fi
# 解密
if [ "$isencrypt" == "true" ]; then
    openssl enc -d -aes-256-cbc -in ${tmp_dir}/payload -out ${tmp_dir}/payload.tar.gz -k "123456"
else
    tail -n+$ARCHIVE "$0" >${tmp_dir}/payload.tar.gz
fi

# 解压
tar -zxf ${tmp_dir}/payload.tar.gz -C $tmp_dir
if [ $? -ne 0 ]; then
    echo "Error: decompose err" >&2
    rm -rf $tmp_dir
    exit 1
fi
cd $tmp_dir
. ./install.sh
cd -
rm -rf $tmp_dir
exit 0

#This line must be the last line of the file
__ARCHIVE_BELOW__
`
