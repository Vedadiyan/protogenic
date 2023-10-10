package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	flaggy "github.com/vedadiyan/flaggy/pkg"
)

type Options struct {
	Type        string   `long:"--type" short:"-t" help:"Generation type (service, api_gateway, client)"`
	Files       []string `long:"--file" short:"-f" help:"The path to the .proto file"`
	Module      string   `long:"--module" short:"-m" help:"The name of the Go module"`
	IncludePath *string  `long:"--include-path" short:"-I" help:"Protoc include path"`
	Help        bool     `long:"--help" short:"-h" help:"Shows help"`
}

func (o Options) Run() {
	if o.Help {
		flaggy.PrintHelp()
		return
	}
	if o.Files == nil {
		RunProtoc()
		return
	}
	if o.Module == "" {
		panic("module is required")
	}
	if o.Type == "" {
		panic("type is required")
	}
	wd, _ := os.Getwd()
	if o.IncludePath != nil {
		wd = *o.IncludePath
	}
	t := o.Type
	if t == "service" {
		t = "service|server"
	}
	if t == "api_gateway" {
		t = "api_gateway|api_gateway_server"
	}
	exec := exec.Command("protoc", "--plugin=protoc-gen-protogenic=./protogenic.exe", strings.Join(o.Files, " "), "--protogenic_out=./", fmt.Sprintf("--protogenic_opt=wd=%s,Module=%s,features=%s", wd, o.Module, t))
	exec.Stdout = os.Stdout
	exec.Stderr = os.Stderr
	err := exec.Run()
	if err != nil {
		panic(err)
	}
}
