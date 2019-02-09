protoc-gen-tsi
==

Generate TypeScript Interface from `Protocol Buffers`.

## Build

```sh
go build plugin/src/protoc-gen-tsi.go
```

## Usage

```sh
protoc --plugin="protoc-gen-tsi=protoc-gen-tsi" --tsi_out="./output" input1.proto input2.proto inputN.proto
```

## License

MIT
