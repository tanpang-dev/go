#!/bin/sh
# 使い方: misc-ppc/tc.sh pkg... (Docker で GOARCH=ppc 向けに型検査)
docker run --rm -v ~/go-ppc:/go-ppc -v go-ppc-cache:/root/.cache golang:1.26 bash -c '
export PATH=/go-ppc/bin:$PATH GOROOT=/go-ppc
cd /go-ppc/misc-ppc && go build -o /tmp/tc typecheck.go && GOOS=linux GOARCH=ppc /tmp/tc '"$*"' 2>&1 | cut -c1-260'
