package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
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
	options := protogen.Options{}
	options.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		params := strings.Split(gen.Request.GetParameter(), ",")
		var module string
		var wd string
		features := make(map[string]bool)
		var featuresRef string
		for _, param := range params {
			parts := strings.Split(param, "=")
			if len(parts) != 2 {
				panic("bad optional parameter")
			}
			if parts[0] == "Module" {
				module = parts[1]
				continue
			}
			if parts[0] == "wd" {
				wd = parts[1]
			}
			if parts[0] == "features" {
				featuresRef = parts[1]
				for _, i := range strings.Split(parts[1], "|") {
					features[i] = true
				}
			}
		}
		for name, f := range gen.FilesByPath {
			_ = wd
			_ = name
			if !f.Generate {
				continue
			}
			if len(f.Services) > 0 {
				if _, ok := features["service"]; ok {
					err := protogenic.GenerateNats(module, gen, f)
					if err != nil {
						panic(err)
					}
				}
				if _, ok := features["api_gateway"]; ok {
					err := protogenic.GenerateAPIGateway(module, gen, f)
					if err != nil {
						panic(err)
					}
				}
				if _, ok := features["server"]; ok {
					err := protogenic.GenerateServer(module, gen, f)
					if err != nil {
						panic(err)
					}
				}
				if _, ok := features["api_gateway_server"]; ok {
					err := protogenic.GenerateAPIGatewayServer(module, gen, f)
					if err != nil {
						panic(err)
					}
				}
				if _, ok := features["client"]; ok {
					err := protogenic.GenerateTypescript(gen, f)
					if err != nil {
						panic(err)
					}
				}
			}
			if _, ok := features["client"]; ok {
				if len(f.Messages) > 0 {
					err := protogenic.GenerateTypescript(gen, f)
					if err != nil {
						panic(err)
					}
				}
				continue
			}
			fileMap := make(map[string]string)
			for i := 0; i < f.Desc.Imports().Len(); i++ {
				file := f.Desc.Imports().Get(i)
				if file.Package().Name() == "protobuf" {
					continue
				}
				options := file.Options().(*descriptorpb.FileOptions)
				goPackage := options.GetGoPackage()
				exec := exec.Command(protogenic.CombinePath(path, wd, protogenicFileName), "-f", file.Path(), "-m", module, "-t", featuresRef)
				exec.Stderr = os.Stderr
				exec.Stdout = os.Stdout
				err := exec.Run()
				if err != nil {
					panic(err)
				}
				fileMap[file.Path()] = goPackage
			}
			// goPackage := f.GoImportPath.String()
			// goPath := strings.ReplaceAll(goPackage, "\"", "")
			// goPath = strings.ReplaceAll(goPath, "__$PATH$__", "")
			// goPath = strings.TrimPrefix(goPath, "/")
			exec := exec.Command("protoc", fmt.Sprintf("--go_out=%s", module), fmt.Sprintf("--proto_path=%s", wd), name)
			exec.Stderr = os.Stderr
			exec.Stdout = os.Stdout
			err := exec.Run()
			if err != nil {
				panic(err)
			}
			fileName := strings.Split(strings.ReplaceAll(name, "\\", "/"), "/")
			finalFileName := protogenic.CombinePath(path, wd, module, strings.ReplaceAll(f.GoImportPath.String(), "\"", ""), fmt.Sprintf("%s.pb.go", strings.ReplaceAll(fileName[len(fileName)-1], ".proto", "")))
			_ = finalFileName
			file, err := os.ReadFile(finalFileName)
			if err != nil {
				panic(err)
			}
			fileStr := string(file)
			for _, value := range fileMap {
				if strings.TrimLeft(value, "/") == strings.ReplaceAll(f.GoImportPath.String(), "\"", "") {
					pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^.*%s.*$\n`, strings.TrimLeft(value, "/")))
					lines := pattern.FindAllString(fileStr, 1)
					if lines == nil {
						continue
					}
					line := lines[0]
					line = strings.Split(strings.TrimLeft(line, "\t"), " ")[0]
					line = strings.TrimRight(line, " ")
					fileStr = pattern.ReplaceAllString(fileStr, "")
					fileStr = strings.ReplaceAll(fileStr, fmt.Sprintf("%s.", line), "")
					continue
				}
				importPath := fmt.Sprintf("%s/%s", strings.TrimRight(module, "/"), strings.TrimLeft(value, "/"))
				fileStr = strings.ReplaceAll(fileStr, value, importPath)
			}
			// path := protogenic.CombinePath(module, strings.ReplaceAll(string(f.GoImportPath), "\"", ""))
			// err = os.MkdirAll(path, os.ModePerm)
			// if err != nil {
			// 	panic(err)
			// }
			// err = os.RemoveAll(protogenic.CombinePath(wd, strings.Split(goPath, "/")[0]))
			// if err != nil {
			// 	panic(err)
			// }
			err = os.WriteFile(finalFileName, []byte(fileStr), os.ModePerm)
			if err != nil {
				panic(err)
			}
			// }
		}
		return nil
	})
}

func Fixer(wd string, path string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s/%s", wd, path), "\\", "/")
}
