package protogenic

import (
	"bytes"
	"fmt"
	"os"
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
	//go:embed templates/nats/service.go.tmpl
	_service string
	//go:embed templates/nats/genql.go.tmpl
	_genql string
	//go:embed templates/nats/postgresql.go.tmpl
	_postgresql string
)

type Callback struct {
	OnSuccess []string
	OnError   []string
}

type Nats struct {
	Callback
	ImportPath               string
	ConnName                 string
	Namespace                string
	Queue                    string
	Url                      string
	Method                   string
	AuthService              string
	AuthServiceCacheInterval int32
	RequestType              string
	ResponseType             string
	RequestMapper            string
	ResponseMapper           string
	CacheInterval            int64
	WebHeaderCollection      map[string]string
	MethodName               string

	ProtogenicVersion string
	CompilerVersion   string
	File              string
}

type GENQL struct {
	Callback
	ImportPath    string
	ConnName      string
	Namespace     string
	Queue         string
	RequestType   string
	ResponseType  string
	Query         string
	CacheInterval int
	MethodName    string

	ProtogenicVersion string
	CompilerVersion   string
	File              string
}

type PostgreSQL struct {
	Callback
	ImportPath     string
	ExtraImports   []string
	ConnName       string
	Dsn            string
	Type           string
	Sql            string
	Namespace      string
	Queue          string
	RequestType    string
	ResponseType   string
	RequestMapper  string
	ResponseMapper string
	CacheInterval  int64
	MethodName     string

	ProtogenicVersion string
	CompilerVersion   string
	File              string
}

func GenerateNats(moduleName string, plugin *protogen.Plugin, file *protogen.File) error {
	pathPrefix, _ := GetPathAndExecutable()
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	serviceTemplate, err := template.New("service").Funcs(_funcs).Parse(_service)
	if err != nil {
		return err
	}
	postgresqlTemplate, err := template.New("postgres").Funcs(_funcs).Parse(_postgresql)
	if err != nil {
		return err
	}
	genqlTemplate, err := template.New("genql").Funcs(_funcs).Parse(_genql)
	if err != nil {
		return err
	}
	for _, service := range file.Services {
		serviceOptions := service.Desc.Options().(*descriptorpb.ServiceOptions)
		nats := proto.GetExtension(serviceOptions, rpc.E_Nats).(*rpc.NATS)
		for _, method := range service.Methods {
			methodOptions := method.Desc.Options().(*descriptorpb.MethodOptions)
			defnition := proto.GetExtension(methodOptions, rpc.E_Definition).(*rpc.Definition)
			rpcOptions := proto.GetExtension(methodOptions, rpc.E_RpcOptions).(*rpc.RpcOptions)
			_ = rpcOptions
			switch defnition.Definition.(type) {
			case *rpc.Definition_Http:
				{
					http := defnition.GetHttp()
					requestMapper := "[]byte{}"
					if http.RequestMapper.GetSql() != "" {
						requestMapper = StringToGoByteArray(http.RequestMapper.GetSql())
					} else if http.RequestMapper.GetFile() != "" {
						file, err := os.ReadFile(CombinePath(pathPrefix, path, http.RequestMapper.GetFile()))
						if err != nil {
							return err
						}
						requestMapper = StringToGoByteArray(string(file))
					}
					responseMapper := "[]byte{}"
					if http.ResponseMapper.GetSql() != "" {
						responseMapper = StringToGoByteArray(http.ResponseMapper.GetSql())
					} else if http.ResponseMapper.GetFile() != "" {
						file, err := os.ReadFile(CombinePath(pathPrefix, path, http.ResponseMapper.GetFile()))
						if err != nil {
							return err
						}
						responseMapper = StringToGoByteArray(string(file))
					}
					global.Register(global.NATS, nats.Connection)
					natsService := Nats{
						ImportPath:               string(file.GoPackageName),
						ConnName:                 nats.Connection,
						Namespace:                strings.ToLower(fmt.Sprintf("%s.%s", nats.Namespace, method.GoName)),
						Queue:                    EmptyIfNill(nats.Queue),
						Url:                      http.Url,
						Method:                   http.Method,
						AuthService:              EmptyIfNill(http.AuthorizationService),
						AuthServiceCacheInterval: IfNill(http.AuthorizationCacheSeconds, -1),
						RequestType:              method.Input.GoIdent.GoName,
						ResponseType:             method.Output.GoIdent.GoName,
						RequestMapper:            requestMapper,
						ResponseMapper:           responseMapper,
						WebHeaderCollection:      make(map[string]string),
						MethodName:               method.GoName,
						CacheInterval:            IfNill(rpcOptions.Configure.CacheInterval, -1),
						Callback: Callback{
							OnSuccess: IfNill(IfNill(rpcOptions, rpc.RpcOptions{}).Events, rpc.RpcOptions_Events{}).OnSuccess,
							OnError:   IfNill(IfNill(rpcOptions, rpc.RpcOptions{}).Events, rpc.RpcOptions_Events{}).OnFailure,
						},
						ProtogenicVersion: GetVersion(),
						CompilerVersion:   plugin.Request.CompilerVersion.String(),
						File:              file.GoImportPath.String(),
					}
					for _, webHeader := range http.Header {
						natsService.WebHeaderCollection[strings.ToLower(webHeader.Key)] = webHeader.Value
					}
					path := strings.Split(strings.ReplaceAll(file.GeneratedFilenamePrefix, "\\", "/"), "/")
					filename := moduleName + "/" + strings.Join(path[:len(path)-1], "/") + "/" + "service.http.pb.go"
					svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
					var serverCode bytes.Buffer
					err := serviceTemplate.Execute(&serverCode, natsService)
					if err != nil {
						return err
					}
					svc.P(serverCode.String())
				}
			case *rpc.Definition_Postgresql:
				{
					postgresql := defnition.GetPostgresql()
					requestMapper := "[]byte{}"
					if postgresql.RequestMapper.GetSql() != "" {
						requestMapper = StringToGoByteArray(postgresql.RequestMapper.GetSql())
					} else if postgresql.RequestMapper.GetFile() != "" {
						file, err := os.ReadFile(CombinePath(pathPrefix, path, postgresql.RequestMapper.GetFile()))
						if err != nil {
							return err
						}
						requestMapper = StringToGoByteArray(string(file))
					}
					responseMapper := "[]byte{}"
					if postgresql.ResponseMapper.GetSql() != "" {
						responseMapper = StringToGoByteArray(postgresql.ResponseMapper.GetSql())
					} else if postgresql.ResponseMapper.GetFile() != "" {
						file, err := os.ReadFile(CombinePath(pathPrefix, path, postgresql.ResponseMapper.GetFile()))
						if err != nil {
							return err
						}
						responseMapper = StringToGoByteArray(string(file))
					}
					sql := "[]byte{}"
					var _type string
					switch postgresql.Sql.(type) {
					case *rpc.PostgreSQL_Command:
						{
							_type = "command"
							switch postgresql.GetCommand().Mapper.(type) {
							case *rpc.Mapper_File:
								{
									file, err := os.ReadFile(CombinePath(pathPrefix, path, postgresql.GetCommand().GetFile()))
									if err != nil {
										return err
									}
									sql = StringToGoByteArray(string(file))
								}
							case *rpc.Mapper_Sql:
								{
									sql = StringToGoByteArray(postgresql.GetCommand().GetSql())
								}
							}
						}
					case *rpc.PostgreSQL_Query:
						{
							_type = "query"
							switch postgresql.GetQuery().Mapper.(type) {
							case *rpc.Mapper_File:
								{
									file, err := os.ReadFile(CombinePath(pathPrefix, path, postgresql.GetQuery().GetFile()))
									if err != nil {
										return err
									}
									sql = StringToGoByteArray(string(file))
								}
							case *rpc.Mapper_Sql:
								{
									sql = StringToGoByteArray(postgresql.GetQuery().GetSql())
								}
							}
						}
					}
					extraImports := make([]string, 0)
					inputImportPathArray := strings.Split(strings.TrimPrefix(strings.TrimSuffix(method.Input.GoIdent.GoImportPath.String(), "\""), "\""), "/")
					var inputPrefix string
					if file.GoImportPath.String() != method.Input.GoIdent.GoImportPath.String() {
						extraImports = append(extraImports, method.Input.GoIdent.GoImportPath.String())
						inputPrefix = fmt.Sprintf("%s.", inputImportPathArray[len(inputImportPathArray)-1])
					}
					global.Register(global.NATS, nats.Connection)
					global.Register(global.POSTGRES, postgresql.GetDsn())
					postgresService := PostgreSQL{
						ImportPath:     string(file.GoPackageName),
						ExtraImports:   extraImports,
						ConnName:       nats.Connection,
						Dsn:            postgresql.GetDsn(),
						Sql:            sql,
						Type:           _type,
						Namespace:      strings.ToLower(fmt.Sprintf("%s.%s", nats.Namespace, method.GoName)),
						Queue:          EmptyIfNill(nats.Queue),
						RequestType:    fmt.Sprintf("%s%s", inputPrefix, method.Input.GoIdent.GoName),
						ResponseType:   method.Output.GoIdent.GoName,
						RequestMapper:  requestMapper,
						ResponseMapper: responseMapper,
						CacheInterval:  IfNill(rpcOptions.Configure.CacheInterval, -1),
						MethodName:     method.GoName,
						Callback: Callback{
							OnSuccess: IfNill(IfNill(rpcOptions, rpc.RpcOptions{}).Events, rpc.RpcOptions_Events{}).OnSuccess,
							OnError:   IfNill(IfNill(rpcOptions, rpc.RpcOptions{}).Events, rpc.RpcOptions_Events{}).OnFailure,
						},
						ProtogenicVersion: GetVersion(),
						CompilerVersion:   plugin.Request.CompilerVersion.String(),
						File:              file.GoImportPath.String(),
					}
					path := strings.Split(strings.ReplaceAll(file.GeneratedFilenamePrefix, "\\", "/"), "/")
					filename := moduleName + "/" + strings.Join(path[:len(path)-1], "/") + "/" + "service.postgres.pb.go"
					svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
					var serverCode bytes.Buffer
					err := postgresqlTemplate.Execute(&serverCode, postgresService)
					if err != nil {
						return err
					}
					svc.P(serverCode.String())
				}
			case *rpc.Definition_Genql:
				{
					http := defnition.GetGenql()
					query := "[]byte{}"
					if http.Query.GetSql() != "" {
						query = StringToGoByteArray(http.Query.GetSql())
					} else if http.Query.GetFile() != "" {
						file, err := os.ReadFile(CombinePath(pathPrefix, path, http.Query.GetFile()))
						if err != nil {
							return err
						}
						query = StringToGoByteArray(string(file))
					}
					global.Register(global.NATS, nats.Connection)
					genqlService := GENQL{
						ImportPath:   string(file.GoPackageName),
						ConnName:     nats.Connection,
						Namespace:    strings.ToLower(fmt.Sprintf("%s.%s", nats.Namespace, method.GoName)),
						Queue:        EmptyIfNill(nats.Queue),
						RequestType:  method.Input.GoIdent.GoName,
						ResponseType: method.Output.GoIdent.GoName,
						Query:        query,
						MethodName:   method.GoName,
						Callback: Callback{
							OnSuccess: IfNill(IfNill(rpcOptions, rpc.RpcOptions{}).Events, rpc.RpcOptions_Events{}).OnSuccess,
							OnError:   IfNill(IfNill(rpcOptions, rpc.RpcOptions{}).Events, rpc.RpcOptions_Events{}).OnFailure,
						},
						ProtogenicVersion: GetVersion(),
						CompilerVersion:   plugin.Request.CompilerVersion.String(),
						File:              file.GoImportPath.String(),
					}

					path := strings.Split(strings.ReplaceAll(file.GeneratedFilenamePrefix, "\\", "/"), "/")
					filename := moduleName + "/" + strings.Join(path[:len(path)-1], "/") + "/" + "service.genql.pb.go"
					svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
					var serverCode bytes.Buffer
					err := genqlTemplate.Execute(&serverCode, genqlService)
					if err != nil {
						return err
					}
					svc.P(serverCode.String())
				}
			default:
				{
					panic("unsupported definition option")
				}
			}
		}
	}
	return nil
}
