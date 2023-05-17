package protogenic

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	rpc "github.com/vedadiyan/protogenic/internal/autogen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	_ "embed"
)

var (
	//go:embed templates/apigateway/aggregated.go.tmpl
	_aggregatedApiGateway string
	//go:embed templates/apigateway/loose.go.tmpl
	_looseApiGateway string
)

type Gateway struct {
	RequestType  string
	ResponseType string
	Namespace    string
}

type LooseAPIGatewayContext struct {
	ImportPath string
	ConnName   string
	Route      string
	Method     string
	Gateways   map[string]Gateway

	ProtogenicVersion string
	CompilerVersion   string
	File              string
}

type AggregatedAPIGatewayContext struct {
	ImportPath   string
	ConnName     string
	Route        string
	Method       string
	RequestType  string
	ResponseType string
	Gateways     map[string]Gateway

	ProtogenicVersion string
	CompilerVersion   string
	File              string
}

func GenerateAPIGateway(plugin *protogen.Plugin, file *protogen.File) error {
	mapper := make(map[string]protoreflect.ProtoMessage)
	var a rpc.APIGateway
	mapper[""] = &a
	for _, service := range file.Services {
		serviceOptions := service.Desc.Options().(*descriptorpb.ServiceOptions)
		nats := proto.GetExtension(serviceOptions, rpc.E_Nats).(*rpc.NATS)
		apiGateway := proto.GetExtension(serviceOptions, rpc.E_ApiGateway).(*rpc.APIGateway)
		if apiGateway.GetEnableAggregation() {
			apiGatewayTemplate, err := template.New("aggregatedapigateway").Funcs(_funcs).Parse(_aggregatedApiGateway)
			if err != nil {
				return err
			}
			gateways := make(map[string]Gateway)
			requests := make(map[string]struct{})
			responses := make(map[string]struct{})
			for _, method := range service.Methods {
				gateway := Gateway{
					Namespace:    fmt.Sprintf("%s.%s", nats.Namespace, strings.ToLower(method.GoName)),
					RequestType:  method.Input.GoIdent.GoName,
					ResponseType: method.Output.GoIdent.GoName,
				}
				gateways[method.GoName] = gateway
				requests[method.Input.GoIdent.GoName] = struct{}{}
				responses[method.Output.GoIdent.GoName] = struct{}{}
			}
			if len(requests) > 1 {
				panic("all requests in an aggregated gateway should be of the same type")
			}
			if len(responses) > 1 {
				panic("all responses in an aggregated should of the same type")
			}
			var requestType string
			var responseType string
			for key := range requests {
				requestType = key
			}
			for key := range responses {
				responseType = key
			}
			gateway := AggregatedAPIGatewayContext{
				ImportPath:   string(file.GoPackageName),
				ConnName:     nats.Connection,
				Route:        apiGateway.Route,
				Method:       IfNill(apiGateway.Method, "GET"),
				RequestType:  requestType,
				ResponseType: responseType,
				Gateways:     gateways,

				ProtogenicVersion: GetVersion(),
				CompilerVersion:   plugin.Request.CompilerVersion.String(),
				File:              file.GoImportPath.String(),
			}
			filename := file.GeneratedFilenamePrefix + fmt.Sprintf("_%s_gateway.pb.go", service.GoName)
			svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
			var serverCode bytes.Buffer
			err = apiGatewayTemplate.Execute(&serverCode, gateway)
			if err != nil {
				return err
			}
			svc.P(serverCode.String())
			continue
		}
		apiGatewayTemplate, err := template.New("looseapigateway").Funcs(_funcs).Parse(_looseApiGateway)
		if err != nil {
			return err
		}
		gateways := make(map[string]Gateway)
		for _, method := range service.Methods {
			gateway := Gateway{
				Namespace:    fmt.Sprintf("%s.%s", nats.Namespace, strings.ToLower(method.GoName)),
				RequestType:  method.Input.GoIdent.GoName,
				ResponseType: method.Output.GoIdent.GoName,
			}
			gateways[method.GoName] = gateway
		}
		gateway := LooseAPIGatewayContext{
			ImportPath: string(file.GoPackageName),
			ConnName:   nats.Connection,
			Route:      apiGateway.Route,
			Method:     IfNill(apiGateway.Method, "GET"),
			Gateways:   gateways,

			ProtogenicVersion: GetVersion(),
			CompilerVersion:   plugin.Request.CompilerVersion.String(),
			File:              file.GoImportPath.String(),
		}
		filename := file.GeneratedFilenamePrefix + fmt.Sprintf("_%s_gateway.pb.go", service.GoName)
		svc := plugin.NewGeneratedFile(strings.ToLower(filename), file.GoImportPath)
		var serverCode bytes.Buffer
		err = apiGatewayTemplate.Execute(&serverCode, gateway)
		if err != nil {
			return err
		}
		svc.P(serverCode.String())
	}
	return nil
}
