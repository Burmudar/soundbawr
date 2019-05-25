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

BOLD := "\\e[1m"
GREEN := "\\e[32m"
RESET := "\\e[0m"

NANOPB_FILENAME := nanopb-0.3.9.3-${PLATFORM}-x86.tar.gz

ARDUINO_DIR := ${CWD}/device
ARDUINO_PROTOBUF_DIR := ${ARDUINO_DIR}/lib/device
SERVICE_DIR := ${CWD}/service
SERVICE_PROTOBUF_DIR := ${SERVICE_DIR}/device

HAS_PROTOBUF := $(shell which protoc > /dev/null && echo $?)

download-nanopb: 
	@echo "downlload nanopb to ${NANOPB_FILENAME}"
	@curl -L "https://jpa.kapsi.fi/nanopb/download/${NANOPB_FILENAME}" -o ${NANOPB_FILENAME}

extract-nanopb: 
ifneq ($(wildcard nanopb),)
	@rm -rf nanopb
endif
	@tar -xf ${NANOPB_FILENAME}
	@mv $(subst .tar.gz,,${NANOPB_FILENAME}) nanopb
	@echo "extracted ${NANOPB_FILENAME} to nanopb"

install-nanopb: download-nanopb extract-nanopb

install-protobuf-macosx:
	@echo "${BOLD}${GREEN}installing protobuf via Homebrew${RESET}"
	@brew install protobuf
	@echo "${BOLD}${GREEN}done${RESET}"

install-protobuf-linux:
	@echo "${BOLD}${GREEN}installing protobuf via apt${RESET}"
	@sudo apt install protobuf-compiler
	@echo "${BOLD}${GREEN}done${RESET}"

install-protobuf:
ifeq ($(HAS_PROTOBUF),0)
	@echo "${BOLD}${GREEN}Protobuf is already installed${RESET}"
else ifeq (${OS},Linux)
	@$(MAKE) install-protobuf-linux
else
	@$(MAKE) install-protobuf-macosx
endif

	@echo "${BOLD}${GREEN}installing proto-gen-go${RESET}"
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@echo "${BOLD}${GREEN}done${RESET}"

protobuf-service:
	@mkdir -p ${SERVICE_PROTOBUF_DIR}
	@protoc --go_out=${SERVICE_PROTOBUF_DIR} Device.proto

protobuf-device:
	@protoc --plugin=protoc-gen-nanopb=nanopb/generator/protoc-gen-nanopb --nanopb_out=${ARDUINO_PROTOBUF_DIR} Device.proto

protobuf: install-protobuf install-nanopb
	@echo "${BOLD}${GREEN}building protobuf spec for service${RESET}"
	$(MAKE) protobuf-service
	@echo "${BOLD}${GREEN}done${RESET}"
	@echo "${BOLD}${GREEN}building protobuf spec for device${RESET}"
	$(MAKE) protobuf-device
	@echo "${BOLD}${GREEN}done${RESET}"

develop:
	@$(MAKE) device-install-deps
	@$(MAKE) protobuf

build-device:
	$(MAKE) device-build BASE_DIR=${ARDUINO_DIR}

build-service:
	@go build
	

