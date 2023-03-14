package protogenic

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	rpc "github.com/vedadiyan/protogenic/internal/autogen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	_ "embed"
)

var (
	//go:embed templates/apigateway/apigateway.go.tmpl
	_apigateway string
)

type APIGatewayContext struct {
	ImportPath   string
	ConnName     string
	Route        string
	Method       string
	RequestType  string
	ResponseType string
	Gateways     map[string]string
	IsAggregated bool
}

func GenerateAPIGateway(plugin *protogen.Plugin, file *protogen.File) error {
	apiGatewayTemplate, err := template.New("apigateway").Funcs(_funcs).Parse(_apigateway)
	if err != nil {
		return err
	}
	for _, service := range file.Services {
		for _, method := range service.Methods {
			options := method.Desc.Options().(*descriptorpb.MethodOptions)
			nats := proto.GetExtension(options, rpc.E_Nats).(*rpc.NATS)
			http := proto.GetExtension(options, rpc.E_Http).([]*rpc.HTTP)
			apiGateway := proto.GetExtension(options, rpc.E_ApiGateway).(*rpc.APIGateway)
			_ = http

			_ = apiGateway
			gateways := make(map[string]string)
			for _, http := range http {
				gateways[http.Name] = fmt.Sprintf("%s.%s", nats.Namespace, http.Name)
			}
			gateway := APIGatewayContext{
				ImportPath:   string(file.GoImportPath),
				ConnName:     nats.Connection,
				Route:        apiGateway.Route,
				Method:       IfNill(apiGateway.Method, "GET"),
				RequestType:  method.Input.GoIdent.GoName,
				ResponseType: method.Output.GoIdent.GoName,
				Gateways:     gateways,
				IsAggregated: true,
			}
			filename := file.GeneratedFilenamePrefix + fmt.Sprintf("_%s_%s_gateway.pb.go", service.GoName, method.GoName)
			svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
			var serverCode bytes.Buffer
			err := apiGatewayTemplate.Execute(&serverCode, gateway)
			if err != nil {
				return err
			}
			svc.P(serverCode.String())
		}

	}
	return nil
}
