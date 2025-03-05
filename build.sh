#!/bin/bash
set -e

SRCDIR=/
BUILDDIR=/build

mkdir -p ${BUILDDIR} 2>/dev/null
cd ${BUILDDIR}
echo "Cloning coredns repo..."
git clone https://github.com/coredns/coredns.git

cd coredns
git checkout v1.9.4

echo "Patching plugin config..."
ed plugin.cfg <<EOED
/rewrite:rewrite
a
json:github.com/xinbenlv/coredns-json
.
w
q
EOED

# Add our module to coredns.
echo "Patching go modules..."
ed go.mod <<EOED
a
replace github.com/xinbenlv/coredns-json => ../../coredns-json
.
/^)
-1
a
	github.com/xinbenlv/coredns-json v0.0.1
.
w
q
EOED

go get github.com/xinbenlv/coredns-json@latest
go get
go mod download

echo "Building..."
make SHELL='sh -x' CGO_ENABLED=1 coredns

cp coredns ${SRCDIR}
cd ${SRCDIR}
rm -r ${BUILDDIR} 