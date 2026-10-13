#!/bin/sh
# Go の main(t.go)を ppc でコンパイルし、偽 runtime とリンクする(M2 確認用)
set -e
cd "$(dirname "$0")"
export GOOS=linux GOARCH=ppc
T=$(go env GOTOOLDIR)
INC=$(go env GOROOT)/pkg/include
$T/compile -p main -o main.o t.go
$T/asm -p runtime -I $INC -o rt.o rt.s
rm -f runtime.a main.a; $T/pack c runtime.a rt.o; $T/pack c main.a main.o
printf 'packagefile runtime=runtime.a\npackagefile main=main.a\n' > importcfg
$T/link -w -importcfg importcfg -E _rt0_ppc_linux -o t main.a
