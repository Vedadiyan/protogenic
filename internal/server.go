package protogenic

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	rpc "github.com/vedadiyan/protogenic/internal/autogen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	_ "embed"
)

var (
	//go:embed templates/server/main.go.tmpl
	_server string
)

type Server struct {
	NatsConns   map[string]string
	UseInfluxDb bool
	InfluxDb    string
	ModuleName  string
	ETCD        string
	Import      string
}

func GenerateServer(moduleName string, plugin *protogen.Plugin, file *protogen.File) error {
	if len(file.Services) == 0 {
		return nil
	}
	serverTemplate, err := template.New("server").Funcs(_funcs).Parse(_server)
	if err != nil {
		return err
	}
	fileOptions := file.Desc.Options().(*descriptorpb.FileOptions)
	etcd := proto.GetExtension(fileOptions, rpc.E_Etcd).(*rpc.ETCD)
	natsConns := make(map[string]string)
	for _, service := range file.Services {
		serviceOptions := service.Desc.Options().(*descriptorpb.ServiceOptions)
		nats := proto.GetExtension(serviceOptions, rpc.E_Nats).(*rpc.NATS)
		natsConns[nats.Connection] = service.GoName
	}
	server := Server{
		NatsConns:   natsConns,
		UseInfluxDb: false,
		ETCD:        etcd.Url,
		Import:      fmt.Sprintf("%s/%s", moduleName, strings.ReplaceAll(string(file.GoImportPath), "\"", "")),
	}
	path := CombinePath(moduleName, "cmd")
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}
	initGo := exec.Command("go", "mod", "init", moduleName)
	initGo.Dir = moduleName
	err = initGo.Run()
	if err != nil {
		panic(err)
	}
	filename := CombinePath(path, "main.go")
	svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
	var serverCode bytes.Buffer
	err = serverTemplate.Execute(&serverCode, server)
	if err != nil {
		return err
	}
	svc.P(serverCode.String())
	return nil
}
