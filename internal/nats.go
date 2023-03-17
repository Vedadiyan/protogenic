package protogenic

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	rpc "github.com/vedadiyan/protogenic/internal/autogen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	_ "embed"
)

var (
	//go:embed templates/nats/service.go.tmpl
	_service string
)

type Nats struct {
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
	CacheInterval            int
	WebHeaderCollection      map[string]string
	MethodName               string
}

type NatsContext struct {
	NatsServices []Nats
	Package      string
}

func GenerateNats(moduleName string, plugin *protogen.Plugin, file *protogen.File) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	serviceTemplate, err := template.New("service").Funcs(_funcs).Parse(_service)
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
			http := defnition.GetHttp()
			switch defnition.Definition.(type) {
			case *rpc.Definition_Http:
				{
					requestMapper := "[]byte{}"
					if http.RequestMapper.GetFile() != "" {
						file, err := os.ReadFile(CombinePath(path, http.RequestMapper.GetFile()))
						if err != nil {
							return err
						}
						requestMapper = StringToGoByteArray(string(file))
					}
					responseMapper := "[]byte{}"
					if http.ResponseMapper.GetFile() != "" {
						file, err := os.ReadFile(CombinePath(path, http.ResponseMapper.GetFile()))
						if err != nil {
							return err
						}
						responseMapper = StringToGoByteArray(string(file))
					}
					natsService := Nats{
						ImportPath:               string(file.GoPackageName),
						ConnName:                 nats.Connection,
						Namespace:                fmt.Sprintf("%s.%s", nats.Namespace, method.GoName),
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
					}
					for _, webHeader := range http.Header {
						natsService.WebHeaderCollection[webHeader.Key] = webHeader.Value
					}
					filename := moduleName + "/" + file.GeneratedFilenamePrefix + fmt.Sprintf("_%s_%s.pb.go", service.GoName, method.GoName)
					svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
					var serverCode bytes.Buffer
					err := serviceTemplate.Execute(&serverCode, natsService)
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
