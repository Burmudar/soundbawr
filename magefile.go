// +build mage

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
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

func progressBar(current, total, segments int64) {
	bytesPerSeg := total / segments
	var currentSeg int64

	if current != int64(0) {
		currentSeg = current / bytesPerSeg
	}

	fmt.Printf("\r[")
	for i := int64(0); i < currentSeg; i++ {
		fmt.Printf("#")
	}

	for i := int64(0); i < segments-currentSeg; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("]")
}

func download(url, filename string) error {
	resp, err := http.Get(url)

	if err != nil {
		log.Printf("Failed to download: %s", url)
		return err
	}

	fmt.Printf("Content-Length: %v\n", resp.ContentLength)

	size := resp.ContentLength
	var i int64
	var total float32
	var now *time.Time
	var window []byte = make([]byte, 64)
	fp, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer fp.Close()
	for i < size {
		n, err := resp.Body.Read(window)
		fp.Write(window[:n])
		if err != nil && err != io.EOF {
			return err
		}

		if now == nil {
			t := time.Now()
			now = &t
		}

		total += float32(len(window))
		i += int64(n)

		if time.Now().Sub(*now).Seconds() >= 1 {
			progressBar(i, size, 100)
			fmt.Printf(" %3.2f MB/s", total/(1024*1024))
			total = 0
			now = nil
		}
	}
	progressBar(i, size, 100)
	fmt.Printf("%3.2f MB/s", total/(1024*1024))
	fmt.Println()
	return nil
}

func DownloadNanoPB() error {
	platform := "linux"

	if runtime.GOOS == "darwin" {
		platform = "macosx"
	}

	filename := fmt.Sprintf("nanopb-0.3.9.3-%s-x86.tar.gz", platform)
	log.Println("download nanopb to %v", filename)

	url := fmt.Sprintf("https://jpa.kapsi.fi/nanopb/download/%s", filename)
	return download(url, filename)
}

func Setup() error {
	return nil
}
