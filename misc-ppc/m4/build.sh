#!/bin/sh
# 公式 Tailscale を linux/ppc(Wii)向けにビルドする。
# 1) tailscale を clone して go mod vendor
# 2) vendor の golang.org/x/sys に ppc 対応を当てる(unix の asm と cpu のビルドタグ)
# 3) --extra-small 相当のタグ + ts_include_cli で tailscaled 1本(CLI 同梱)を作る
set -eu
cd "$(dirname "$0")"
[ -d tailscale ] || git clone -q --depth 1 https://github.com/tailscale/tailscale.git
docker run --rm -v ~/go-ppc:/go-ppc -v go-ppc-cache:/root/.cache -v go-ppc-mod:/go/pkg/mod golang:1.26 bash -c '
set -e
export PATH=/go-ppc/bin:$PATH GOROOT=/go-ppc GOTOOLCHAIN=local
cd /go-ppc/misc-ppc/m4/tailscale
go mod vendor
X=vendor/golang.org/x/sys
cp ../xsys-ppc/asm_linux_ppc.s $X/unix/
sed -i "s#^//go:build mips || mipsle\$#//go:build mips || mipsle || ppc#" $X/cpu/cpu_mipsx.go
T=$(go run -mod=vendor ./cmd/featuretags --extra-small)
GOOS=linux GOARCH=ppc CGO_ENABLED=0 go build -mod=vendor -tags="$T,ts_include_cli" -trimpath -ldflags="-s -w" -o ../tailscaled-small ./cmd/tailscaled
ls -l ../tailscaled-small'
