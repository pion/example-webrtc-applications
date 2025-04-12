$Env:GOOS = 'js'
$Env:GOARCH = 'wasm'
go build -o gopher-combat.wasm valorzard/gopher-combat
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

$goroot = go env GOROOT
# have to copy the wasm_exec.js file to the current directory
# the location of this file changed in go 1.24
cp $goroot\lib\wasm\wasm_exec.js .

# serve the files
python3 -m http.server