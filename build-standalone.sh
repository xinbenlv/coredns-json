#!/bin/bash
set -e

SRCDIR=`pwd`
BUILDDIR=`pwd`/build

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
replace github.com/xinbenlv/coredns-json => ../..
.
/^)
-1
a
	github.com/xinbenlv/coredns-json v0.0.1
.
w
q
EOED

go get github.com/xinbenlv/coredns-json@v0.0.1
go get
go mod download

echo "Building..."
# run make coredns	
make coredns

cp coredns ${SRCDIR}
chmod -R 755 .git
cd ${SRCDIR}
rm -r ${BUILDDIR} 