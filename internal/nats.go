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
	ImportPath          string
	ConnName            string
	Namespace           string
	Queue               string
	Url                 string
	Method              string
	AuthService         string
	RequestType         string
	ResponseType        string
	RequestMapper       string
	ResponseMapper      string
	CacheInterval       int
	WebHeaderCollection map[string]string
}

type NatsContext struct {
	NatsServices []Nats
	Package      string
}

func GenerateNats(plugin *protogen.Plugin, file *protogen.File) error {
	serviceTemplate, err := template.New("service").Funcs(_funcs).Parse(_service)
	if err != nil {
		return err
	}
	for _, service := range file.Services {
		for _, method := range service.Methods {
			options := method.Desc.Options().(*descriptorpb.MethodOptions)
			nats := proto.GetExtension(options, rpc.E_Nats).(*rpc.NATS)
			http := proto.GetExtension(options, rpc.E_Http).([]*rpc.HTTP)
			microservice := proto.GetExtension(options, rpc.E_Microservice).(*rpc.Microservice)
			apiGateway := proto.GetExtension(options, rpc.E_ApiGateway).(*rpc.APIGateway)
			_ = http
			_ = microservice
			_ = apiGateway
			for _, http := range http {
				requestMapper := "[]byte{}"
				if http.RequestMapper.GetFile() != "" {
					file, err := os.ReadFile(http.RequestMapper.GetFile())
					if err != nil {
						return err
					}
					requestMapper = StringToGoByteArray(string(file))
				}
				responseMapper := "[]byte{}"
				if http.ResponseMapper.GetFile() != "" {
					file, err := os.ReadFile(http.ResponseMapper.GetFile())
					if err != nil {
						return err
					}
					responseMapper = StringToGoByteArray(string(file))
				}
				natsService := Nats{
					ImportPath:          string(file.GoImportPath),
					ConnName:            nats.Connection,
					Namespace:           fmt.Sprintf("%s.%s", nats.Namespace, http.Name),
					Queue:               EmptyIfNill(nats.Queue),
					Url:                 http.Url,
					Method:              http.Method,
					AuthService:         EmptyIfNill(http.AuthorizationService),
					RequestType:         method.Input.GoIdent.GoName,
					ResponseType:        method.Output.GoIdent.GoName,
					RequestMapper:       requestMapper,
					ResponseMapper:      responseMapper,
					WebHeaderCollection: make(map[string]string),
				}
				for _, webHeader := range http.Header {
					natsService.WebHeaderCollection[webHeader.Key] = webHeader.Value
				}
				filename := file.GeneratedFilenamePrefix + fmt.Sprintf("_%s_%s_%s.pb.go", service.GoName, method.GoName, http.Name)
				svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
				var serverCode bytes.Buffer
				err := serviceTemplate.Execute(&serverCode, natsService)
				if err != nil {
					return err
				}
				svc.P(serverCode.String())
			}
		}

	}
	return nil
}
