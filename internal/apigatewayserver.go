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
	//go:embed templates/apigateway_server/main.go.tmpl
	_apigatewayserver string
)

type APIGatewayServer struct {
	NatsConns   []string
	ModuleName  string
	UseEtcd     string
	UseMongoDb  string
	UseRedis    string
	UseInfluxDb string
	Import      string
}

func GenerateAPIGatewayServer(moduleName string, plugin *protogen.Plugin, file *protogen.File) error {
	if len(file.Services) == 0 {
		return nil
	}
	serverTemplate, err := template.New("apigatewayserver").Funcs(_funcs).Parse(_apigatewayserver)
	if err != nil {
		return err
	}
	fileOptions := file.Desc.Options().(*descriptorpb.FileOptions)
	useEtcd := proto.GetExtension(fileOptions, rpc.E_UseEtcd).(*string)
	useMongodb := proto.GetExtension(fileOptions, rpc.E_UseMongoDb).(*string)
	useRedis := proto.GetExtension(fileOptions, rpc.E_UseRedis).(*string)
	useInfluxDb := proto.GetExtension(fileOptions, rpc.E_UseInfluxDb).(*string)
	natsConns := make([]string, 0)
	for _, service := range file.Services {
		serviceOptions := service.Desc.Options().(*descriptorpb.ServiceOptions)
		nats := proto.GetExtension(serviceOptions, rpc.E_Nats).(*rpc.NATS)
		natsConns = append(natsConns, nats.Connection)
	}
	server := APIGatewayServer{
		NatsConns:   natsConns,
		UseEtcd:     EmptyIfNill(useEtcd),
		UseMongoDb:  EmptyIfNill(useMongodb),
		UseRedis:    EmptyIfNill(useRedis),
		UseInfluxDb: EmptyIfNill(useInfluxDb),
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
