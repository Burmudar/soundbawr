# Arduino Dependencies
* IRremoteESP8266

# Dependencies
* Protobuf 3.7.1
* https://jpa.kapsi.fi/nanopb/download/nanopb-0.3.9.3-macosx-x86.tar.gz (replace `macosx` with `linux` for linux) - Nanopb

## Mac
* `brew install protobuf`
## Linux
* `sudo apt install protobuf-compiler python-protobuf`

## Generate protobuf files
**Arduino-Remote**: `protoc --plugin=protoc-gen-nanopb=nanopb/generator/protoc-gen-nanopb --nanopb_out=device/lib/device/ Device.proto`
**Server**: `mkdir -p server/device && protoc --go_out=server/device Device.proto"