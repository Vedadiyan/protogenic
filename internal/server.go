package protogenic

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	rpc "github.com/vedadiyan/protogenic/internal/autogen"
	"github.com/vedadiyan/protogenic/internal/global"
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
	NatsConns     []string
	PostgresConns []string
	RedisConns    []string
	MongoConns    []string
	UseInfluxDb   bool
	InfluxDb      string
	ModuleName    string
	ETCD          string
	Import        string
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
	natsConns := make([]string, 0)
	for _, service := range file.Services {
		serviceOptions := service.Desc.Options().(*descriptorpb.ServiceOptions)
		nats := proto.GetExtension(serviceOptions, rpc.E_Nats).(*rpc.NATS)
		natsConns = append(natsConns, nats.Connection)
	}
	postgresConns := make([]string, 0)
	redisConns := make([]string, 0)
	mongoConns := make([]string, 0)
	global.ForEach(func(dependencyType global.DEPENDENCY_TYPES, key string) {
		switch dependencyType {
		case global.POSTGRES:
			{
				postgresConns = append(postgresConns, key)
			}
		case global.REDIS:
			{
				redisConns = append(redisConns, key)
			}
		case global.MONGO:
			{
				mongoConns = append(mongoConns, key)
			}
		}
	})
	server := Server{
		NatsConns:     natsConns,
		PostgresConns: postgresConns,
		RedisConns:    redisConns,
		MongoConns:    mongoConns,
		UseInfluxDb:   false,
		ETCD:          etcd.Url,
		Import:        fmt.Sprintf("%s/%s", moduleName, strings.ReplaceAll(string(file.GoImportPath), "\"", "")),
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
