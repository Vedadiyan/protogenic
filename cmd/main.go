package main

import (
	"fmt"
	"os"
	"os/exec"

	protogenic "github.com/vedadiyan/protogenic/internal"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	wd, _ := os.Getwd()
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		for name, f := range gen.FilesByPath {
			if !f.Generate {
				continue
			}
			err := protogenic.GenerateNats(gen, f)
			if err != nil {
				panic(err)
			}
			err = protogenic.GenerateAPIGateway(gen, f)
			if err != nil {
				panic(err)
			}
			if len(f.Messages) > 0 {
				err = protogenic.GenerateTypescript(gen, f)
				if err != nil {
					panic(err)
				}
				exec := exec.Command("protoc", "--go_out=./", fmt.Sprintf("--proto_path=%s", wd), fmt.Sprintf("%s/%s", wd, name))
				exec.Stderr = os.Stderr
				exec.Stdout = os.Stdout
				err := exec.Run()
				if err != nil {
					panic(err)
				}
			}
		}
		return nil
	})
}
