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

if [ -n "$signature" ] && [ -n "$isencrypt" ] && [ "$isencrypt" == "true" ]; then
    if [ $# -ne 3 ]; then
        echo "Error: The number of parameters is incorrect." >&2
        rm -rf $tmp_dir
        exit 1
    fi
    pubkey=$1
    key=$2
    iv=$3
fi

if [ -n "$signature" ] && [ -n "$isencrypt" ] && [ "$isencrypt" == "false" ]; then
    if [ $# -ne 1 ]; then
        echo "Error: The number of parameters is incorrect." >&2
        rm -rf $tmp_dir
        exit 1
    fi
    pubkey=$1
fi

if [ -z "$signature" ] && [ -n "$isencrypt" ]; then
    if [ $# -ne 2 ]; then
        echo "Error: The number of parameters is incorrect." >&2
        rm -rf $tmp_dir
        exit 1
    fi
    key=$1
    iv=$2
fi

if [ -z "$signature" ] && [ -z "$isencrypt" ]; then
    if [ $# -ne 0 ]; then
        echo "Error: The number of parameters is incorrect." >&2
        rm -rf $tmp_dir
        exit 1
    fi
fi


# 检查签名值是否不为空
if [ -n "$signature" ]; then
    # 创建签名文件
    echo "$signature" >${tmp_dir}/signature
    # 创建负载文件
    tail -n+$ARCHIVE "$0" >${tmp_dir}/payload
    xxd -r -p ${tmp_dir}/signature ${tmp_dir}/signature.bin
    # 使用openssl验证签名
    openssl dgst -verify $pubkey -signature ${tmp_dir}/signature.bin ${tmp_dir}/payload >/dev/null 2>&1
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
    tail -n+$ARCHIVE "$0" >${tmp_dir}/payload
    openssl aes-256-cbc -d -in ${tmp_dir}/payload -out ${tmp_dir}/payload.tar.gz -K $key -iv $iv -nosalt >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "decrypt succeeded."
        rm -rf ${tmp_dir}/payload
    else
        echo "Error: decrypt failed." >&2
        rm -rf $tmp_dir
        exit 1
    fi
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
