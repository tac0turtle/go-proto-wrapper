# Golang Protobuf Msg Wrapper

This tool is meant to generate wrap and unwrap methods for protobuf `oneof`'s.

## Installation

```sh
go get github.com/mabar3778/go-proto-wrapper
```

## Usage

### Buf

With [buf](https://buf.build/) you can use it with the `buf generate` command by adding it to you `buf.gen.yaml` file like so:

```yaml
version: v1beta1
plugins:
  - name: gogo
    out: .
    opt: paths=source_relative
  - name: gowrapper
    out: .
    opt: gogoimport=true
```
