#!/bin/sh
# 標準ライブラリのテストを ppc でビルドし qemu-ppc(750cl)で回す。使い方: run.sh pkg...
# 1パッケージ900秒まで。結果は tests/<pkg>.out
set -u
cd "$(dirname "$0")"
docker run --rm -v ~/go-ppc:/go-ppc -v go-ppc-cache:/root/.cache golang:1.26 bash -c '
export PATH=/go-ppc/bin:$PATH GOROOT=/go-ppc; cd /go-ppc/src
for p in '"$*"'; do n=$(echo $p | tr / _); GOOS=linux GOARCH=ppc go test -c -o /go-ppc/misc-ppc/tests/$n.test $p > /go-ppc/misc-ppc/tests/$n.build 2>&1 || echo "BUILD-FAIL $p"; done'
# 自分自身を exec し直すテストのため、実行中だけ binfmt に qemu-ppc を登録する(Docker VM 共有なので終了時に外す)
docker run --rm --privileged -v ~/go-ppc:/go-ppc ubuntu:24.04 bash -c '
export DEBIAN_FRONTEND=noninteractive QEMU_CPU=750cl; apt-get update -qq >/dev/null; apt-get install -y -qq qemu-user-static >/dev/null 2>&1
B=/proc/sys/fs/binfmt_misc; grep -q " $B " /proc/mounts || mount -t binfmt_misc none $B
[ -e $B/goppc ] || printf ":goppc:M::\\x7fELF\\x01\\x02\\x01\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x02\\x00\\x14:\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\x00\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xfe\\xff\\xff:/usr/bin/qemu-ppc-static:" > $B/register
trap "echo -1 > $B/goppc" EXIT
for p in '"$*"'; do n=$(echo $p | tr / _); [ -f /go-ppc/misc-ppc/tests/$n.test ] || continue
  cd /go-ppc/src/$p && timeout 900 /go-ppc/misc-ppc/tests/$n.test -test.short > /go-ppc/misc-ppc/tests/$n.out 2>&1
  printf "%-28s %s\n" "$p" "$(grep -aE "^(ok|PASS|FAIL)$|^--- FAIL|^panic:|fatal error" /go-ppc/misc-ppc/tests/$n.out | head -3 | tr "\n" " ")"
done'
