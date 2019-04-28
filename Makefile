include device/device.mk
OS := $(shell uname -s)
CWD := $(shell pwd)

ifeq (${OS},Linux)
	PLATFORM := linux
	GOOS := linux
else
	PLATFORM := macosx
    GOOS := darwin
endif

NANOPB_FILENAME := nanopb-0.3.9.3-${PLATFORM}-x86.tar.gz

ARDUINO_DIR := ${CWD}/device
ARDUINO_PROTOBUF_DIR := ${ARDUINO_DIR}/lib/device
SERVICE_DIR := ${CWD}/service
SERVICE_PROTOBUF_DIR := ${SERIVCE_DIR}/device

download-nanopb: ${NANOPB_FILENAME}
	@curl -L "https://jpa.kapsi.fi/nanopb/download/${NANOPB_FILENAME}" -o ${NANOPB_FILENAME}

extract-nanopb: ${NANOPB_FILENAME}
	@tar -xvf ${NANOPB_FILENAME}
	@mv $(subst .tar.gz,,${NANOPB_FILENAME}) nanopb

nanopb: download-nanopb extract-nanopb

install-protobuf-macosx:
	@brew install protobuf

has-protobuf:
	@which protoc > /dev/null

protobuf-service:
	@mkdir -p ${SERVICE_PROTOBUF_DIR}
	@protoc --go_out=${SERVICE_PROTOBUF_DIR} Device.proto

protobuf-device:
	@protoc --plugin=protoc-gen-nanopb=nanopb/generator/protoc-gen-nanopb --nanopb_out=${ARDUINO_PROTOBUF_DIR} Device.proto

protobuf: nanopb
	$(MAKE) protobuf-service
	$(MAKE) protobuf-device

develop:
	$(MAKE) device-install-deps
	$(MAKE) protobuf

build-device:
	$(MAKE) device-build BASE_DIR=${ARDUINO_DIR}

build-service:
	@go build
	

