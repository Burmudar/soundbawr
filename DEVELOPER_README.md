# Mac Dependencies
* Protobuf 3.7.1
* https://jpa.kapsi.fi/nanopb/download/nanopb-0.3.9.3-macosx-x86.tar.gz - Nanopb

Generate protobuf files
**Arduino-Remote**: `protoc --plugin=protoc-gen-nanopb=nanopb/generator/protoc-gen-nanopb --nanopb_out=lib/rdevice/ Device.proto`
**Server**: `mkdir -p server/device && protoc --go_out=server/device Device.proto"