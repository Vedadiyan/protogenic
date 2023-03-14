package main

import (
	protogenic "github.com/vedadiyan/protogenic/internal"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		for _, f := range gen.Files {
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
			err = protogenic.GenerateTypescript(gen, f)
			if err != nil {
				panic(err)
			}
		}
		return nil
	})
}
