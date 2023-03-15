package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Options struct {
	Files       []string `long:"--file" short:"-f" help:"The path to the .proto file"`
	IncludePath *string  `long:"--include-path" short:"-I" help:"Protoc include path"`
}

func (o Options) Run() {
	if o.Files == nil {
		RunProtoc()
		return
	}
	wd, _ := os.Getwd()
	if o.IncludePath != nil {
		wd = *o.IncludePath
	}
	exec := exec.Command("protoc", "--plugin=protoc-gen-protogenic=./protogenic.exe", strings.Join(o.Files, " "), "--protogenic_out=./", fmt.Sprintf("--protogenic_opt=%s", wd))
	exec.Stdout = os.Stdout
	exec.Stderr = os.Stderr
	err := exec.Run()
	if err != nil {
		panic(err)
	}
}