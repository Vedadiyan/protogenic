package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	flaggy "github.com/vedadiyan/flaggy/pkg"
	protogenic "github.com/vedadiyan/protogenic/internal"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

func main() {
	options := Options{}
	err := flaggy.Parse(&options, os.Args[1:])
	if err != nil {
		panic(err)
	}
}

func RunProtoc() {
	wd, _ := os.Getwd()
	options := protogen.Options{}
	options.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		for name, f := range gen.FilesByPath {
			_ = wd
			_ = name
			if !f.Generate {
				continue
			}
			// err := protogenic.GenerateNats(gen, f)
			// if err != nil {
			// 	panic(err)
			// }
			// err = protogenic.GenerateAPIGateway(gen, f)
			// if err != nil {
			// 	panic(err)
			// }
			if len(f.Messages) > 0 {
				// err = protogenic.GenerateTypescript(gen, f)
				// if err != nil {
				// 	panic(err)
				// }
				fileMap := make(map[string]string)
				for i := 0; i < f.Desc.Imports().Len(); i++ {
					file := f.Desc.Imports().Get(i)
					options := file.Options().(*descriptorpb.FileOptions)
					goPackage := options.GetGoPackage()
					exec := exec.Command(protogenic.CombinePath(wd, "protogenic2.exe"), "-f", file.Path())
					exec.Stderr = os.Stderr
					exec.Stdout = os.Stdout
					err := exec.Run()
					if err != nil {
						panic(err)
					}
					fileMap[file.Path()] = goPackage
				}
				goPackage := f.GoImportPath.String()
				goPath := strings.ReplaceAll(goPackage, "\"", "")
				// goPath = strings.ReplaceAll(goPath, "__$PATH$__", "")
				// goPath = strings.TrimPrefix(goPath, "/")
				exec := exec.Command("protoc", "--go_out=.", fmt.Sprintf("--proto_path=%s", wd), name)
				exec.Stderr = os.Stderr
				exec.Stdout = os.Stdout
				err := exec.Run()
				if err != nil {
					panic(err)
				}
				fileName := strings.Split(strings.ReplaceAll(name, "\\", "/"), "/")
				finalFileName := fmt.Sprintf("%s.pb.go", strings.ReplaceAll(fileName[len(fileName)-1], ".proto", ""))
				file, err := os.ReadFile(protogenic.CombinePath(wd, goPath, finalFileName))
				if err != nil {
					panic(err)
				}
				fileStr := string(file)
				for _, value := range fileMap {
					fileStr = strings.ReplaceAll(fileStr, value, fmt.Sprintf("%s/%s", "OKKKKK", value))
				}
				err = os.WriteFile(protogenic.CombinePath(wd, goPath, finalFileName), []byte(fileStr), os.ModePerm)
				if err != nil {
					panic(err)
				}
			}
		}
		return nil
	})
}

func Fixer(wd string, path string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s/%s", wd, path), "\\", "/")
}
