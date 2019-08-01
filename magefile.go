// +build mage

package main

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/magefile/mage/sh"
)

type ProjectPath struct {
	base string
	lib  string
}

func (p *ProjectPath) Root() string {
	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	return path.Join(cwd, p.base)
}

func (p *ProjectPath) ProtobufDir() string {
	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	return path.Join(cwd, p.base, p.lib)
}

var arduinoPaths = &ProjectPath{"device", "lib/device"}
var servicePaths = &ProjectPath{"service", "device"}

var Curl = sh.OutCmd("curl")
var Tar = sh.OutCmd("tar")
var Bash = sh.OutCmd("bash", "-c")
var BrewInstall = sh.OutCmd("brew", "install")
var AptInstall = sh.OutCmd("sudo", "apt", "install")
var GoGet = sh.OutCmd("go", "get", "-u")
var Protoc = sh.OutCmd("protoc")

func isMac() bool {
	return runtime.GOOS == "darwin"
}

func platform() string {
	if isMac() {
		return "macosx"
	}
	return "linux"
}

func downloadNanoPB() (string, error) {

	filename := fmt.Sprintf("nanopb-0.3.9.3-%s-x86.tar.gz", platform())
	fmt.Printf("download nanopb to %v\n", filename)

	url := fmt.Sprintf("https://jpa.kapsi.fi/nanopb/download/%s", filename)

	_, err := Curl(url, "-o", filename)
	return filename, err
}

func extract(filename string) (string, error) {
	return Tar("-x", "-v", "-f", filename)
}

func install(pkg string) (string, error) {
	var installFn func(args ...string) (string, error) = AptInstall
	if isMac() {
		installFn = BrewInstall
	}
	fmt.Printf("installing [%v]\n", pkg)
	output, err := installFn(pkg)
	fmt.Printf("done.\n")
	return output, err
}

func installProtobuf() {
	_, err := Bash("which protoc > /dev/null")
	if sh.ExitStatus(err) == 0 {
		fmt.Printf("protobuf already installed\n")
	} else if isMac() {
		install("protobuf")
	} else {
		install("protobuf-compiler")
	}

	fmt.Printf("install proto-gen-go\n")
	GoGet("github.com/golang/protobuf/protoc-gen-go")
	fmt.Printf("done.\n")
}

func GenerateServiceProtobuf() (string, error) {
	sh.Run("mkdir", "-p", servicePaths.ProtobufDir())
	dest := fmt.Sprintf("--go_out=%s", servicePaths.ProtobufDir())
	fmt.Printf("Generating service protobuf files to: %s", servicePaths.ProtobufDir())
	return Protoc(dest, "Device.proto")
}

func GenerateDeviceProtoBuf() (string, error) {
	pluginArg := "--plugin=protoc-gen-nanopb=nanopb/generator/protoc-gen-nanopb"
	dest := fmt.Sprintf("--nanopb_out=%s", arduinoPaths.ProtobufDir())
	fmt.Printf("Generating device protobuf files to: %s", arduinoPaths.ProtobufDir())
	return Protoc(pluginArg, dest, "Device.proto")
}

func Protobuf() {
	installProtobuf()
	out, _ := GenerateServiceProtobuf()
	fmt.Println(out)
	out, _ = GenerateDeviceProtoBuf()
	fmt.Println(out)
}

func Setup() error {
	return nil
}
