# SPDX-FileCopyrightText: 2025 The Pion community <https://pion.ly>
# SPDX-License-Identifier: MIT

$Env:GOOS = 'js'
$Env:GOARCH = 'wasm'
go build -o game.wasm .
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
$goroot = go env GOROOT

# Go 1.24 and newer
cp $goroot\lib\wasm\wasm_exec.js .