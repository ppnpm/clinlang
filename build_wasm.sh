#!/bin/bash
set -e

# Get the Go Root path and strip invisible Windows characters
RAW_GOROOT=$(go env GOROOT | tr -d '\r\n')

# Convert "C:\Program Files\Go" format to Git Bash friendly "/c/Program Files/Go" format
BASH_GOROOT="/$(echo $RAW_GOROOT | sed -e 's/://' -e 's/\\/\//g')"

echo "Copying from: $BASH_GOROOT/lib/wasm/wasm_exec.js"
cp "$BASH_GOROOT/lib/wasm/wasm_exec.js" web/

echo "Compiling clinlang.wasm..."
export GOOS=js
export GOARCH=wasm
go build -o web/clinlang.wasm ./cmd/wasm

echo "Successfully built WebAssembly engine!"
