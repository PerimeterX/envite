#!/bin/bash

set -e

if which go-bindata >/dev/null; then
    echo "building..."
else
    echo "cannot find go-bindata. to install it, run 'go install github.com/go-bindata/go-bindata/v3/go-bindata@v3.1.3'"
    exit 1
fi

npm --prefix ui run build

go-bindata -o static_files.go -pkg envite -prefix ui/build -fs ui/build/...
