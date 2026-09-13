# tailcat

`filecat photo.jpg` で、Tailcat経由の一時的なファイル転送を行うCLIです。

受信ページ: https://sugurunatsuno.github.io/tailcat/

## Status

MVPの実装を開始した段階です。

## Build

```sh
go build ./cmd/filecat
GOOS=js GOARCH=wasm go build -o web/filecat.wasm ./wasm
```
