#!/bin/bash

set -eu -o pipefail

tmp=$(realpath "$0")
dir=$(dirname "$tmp")
cd "$dir"

rm -f find-ami.zip
tmpdir=$(mktemp -d ./pkg-XXXXXX)
pip3 install --system --target "$tmpdir" -r requirements.txt
cd "$tmpdir"
zip -r9 ../find-ami.zip .
cd ..
rm -rf "$tmpdir"
zip -g find-ami.zip find-ami.py
