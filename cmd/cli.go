package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	flaggy "github.com/vedadiyan/flaggy/pkg"
)

type Serve struct {
	host        string `long:"--host" short:"-h" help="Hostname"`
	serviceName string `long:"--service-name" short:"-sn" help:"Service Name"`
	etcdUrl     string `long="--etcd-url" short:"" help="ETCD URL"`
}

func (s Serve) Run() error {

	return nil
}

type Options struct {
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
	wd, _ := os.Getwd()
	if o.IncludePath != nil {
		wd = *o.IncludePath
	}
	exec := exec.Command("protoc", "--plugin=protoc-gen-protogenic=./protogenic.exe", strings.Join(o.Files, " "), "--protogenic_out=./", fmt.Sprintf("--protogenic_opt=wd=%s,Module=%s,features=service|server", wd, o.Module))
	exec.Stdout = os.Stdout
	exec.Stderr = os.Stderr
	err := exec.Run()
	if err != nil {
		panic(err)
	}
}
