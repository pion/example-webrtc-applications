# SPDX-FileCopyrightText: 2025 The Pion community <https://pion.ly>
# SPDX-License-Identifier: MIT

# we have to do this for Go 1.24 and newer
# see: https://tip.golang.org/doc/go1.24#wasm
cp $(go env GOROOT)/lib/wasm/wasm_exec.js .

env GOOS=js GOARCH=wasm go build -o game.wasm