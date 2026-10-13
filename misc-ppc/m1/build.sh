#!/bin/sh
# 偽 runtime + 空の main を組み立てて linux/ppc 実行ファイルにする(M1 確認用)
set -e
cd "$(dirname "$0")"
export GOOS=linux GOARCH=ppc
T=$(go env GOTOOLDIR)
INC=$(go env GOROOT)/pkg/include
$T/asm -p runtime -I $INC -o rt.o rt.s
$T/asm -p main -I $INC -o main.o main.s
rm -f runtime.a main.a; $T/pack c runtime.a rt.o
# アセンブラの出力には compile が書く "main" 行が無く main と認識されないので、ヘッダ直後に差し込む
python3 -c 'import sys;b=open("main.o","rb").read();i=b.index(b"\n")+1;open("main.o","wb").write(b[:i]+b"main\n"+b[i:])' 2>/dev/null || perl -0pi -e 's/\n/\nmain\n/' main.o
$T/pack c main.a main.o
printf 'packagefile runtime=runtime.a\npackagefile main=main.a\n' > importcfg
$T/link -w -importcfg importcfg -E _rt0_ppc_linux -o hello main.a
