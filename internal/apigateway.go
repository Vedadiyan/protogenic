package protogenic

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	rpc "github.com/vedadiyan/protogenic/internal/autogen"
	"github.com/vedadiyan/strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
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
	ImportPath   string
	ExtraImports []string
	ConnName     string
	Route        string
	Method       string
	Gateways     map[string]Gateway

	ProtogenicVersion string
	CompilerVersion   string
	File              string
	UseMeta           bool
	UseValidation     bool
}

type AggregatedAPIGatewayContext struct {
	ImportPath   string
	ExtraImports []string
	ConnName     string
	Route        string
	Method       string
	RequestType  string
	ResponseType string
	Gateways     map[string]Gateway

	ProtogenicVersion string
	CompilerVersion   string
	File              string
	UseMeta           bool
	UseValidation     bool
}

func GenerateAPIGateway(moduleName string, plugin *protogen.Plugin, file *protogen.File) error {
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
			extraImports := make(map[string]bool, 0)
			for _, method := range service.Methods {
				inputImportPathArray := strings.Split(strings.TrimPrefix(strings.TrimSuffix(method.Input.GoIdent.GoImportPath.String(), "\""), "\""), "/")
				var inputPrefix string
				if file.GoImportPath.String() != method.Input.GoIdent.GoImportPath.String() {
					extraImports[method.Input.GoIdent.GoImportPath.String()] = true
					inputPrefix = fmt.Sprintf("%s.", inputImportPathArray[len(inputImportPathArray)-1])
				}
				gateway := Gateway{
					Namespace:    fmt.Sprintf("%s.%s", nats.Namespace, strings.ToLower(method.GoName)),
					RequestType:  fmt.Sprintf("%s%s", inputPrefix, method.Input.GoIdent.GoName),
					ResponseType: method.Output.GoIdent.GoName,
				}
				gateways[strcase.ToSnake(method.GoName)] = gateway
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
			_extraImports := make([]string, len(extraImports))
			for key := range extraImports {
				_extraImports = append(_extraImports, key)
			}
			gateway := AggregatedAPIGatewayContext{
				ImportPath:   string(file.GoPackageName),
				ExtraImports: _extraImports,
				ConnName:     nats.Connection,
				Route:        apiGateway.Route,
				Method:       IfNill(apiGateway.Method, "GET"),
				RequestType:  requestType,
				ResponseType: responseType,
				Gateways:     gateways,

				ProtogenicVersion: GetVersion(),
				CompilerVersion:   plugin.Request.CompilerVersion.String(),
				File:              file.GoImportPath.String(),
				UseMeta:           IfNill(apiGateway.UseMeta, false),
				UseValidation:     IfNill(apiGateway.UseValidation, false),
			}
			path := strings.Split(strings.ReplaceAll(file.GeneratedFilenamePrefix, "\\", "/"), "/")
			filename := moduleName + "/" + strings.Join(path[:len(path)-1], "/") + "/" + "gateway.pb.go"
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
		extraImports := make(map[string]bool)
		for _, method := range service.Methods {
			inputImportPathArray := strings.Split(strings.TrimPrefix(strings.TrimSuffix(method.Input.GoIdent.GoImportPath.String(), "\""), "\""), "/")
			var inputPrefix string
			if file.GoImportPath.String() != method.Input.GoIdent.GoImportPath.String() {
				extraImports[method.Input.GoIdent.GoImportPath.String()] = true
				inputPrefix = fmt.Sprintf("%s.", inputImportPathArray[len(inputImportPathArray)-1])
			}
			gateway := Gateway{
				Namespace:    fmt.Sprintf("%s.%s", nats.Namespace, strings.ToLower(method.GoName)),
				RequestType:  fmt.Sprintf("%s%s", inputPrefix, method.Input.GoIdent.GoName),
				ResponseType: method.Output.GoIdent.GoName,
			}
			gateways[method.GoName] = gateway
		}
		_extraImports := make([]string, len(extraImports))
		for key := range extraImports {
			_extraImports = append(_extraImports, key)
		}
		gateway := LooseAPIGatewayContext{
			ImportPath:   string(file.GoPackageName),
			ExtraImports: _extraImports,
			ConnName:     nats.Connection,
			Route:        apiGateway.Route,
			Method:       IfNill(apiGateway.Method, "GET"),
			Gateways:     gateways,

			ProtogenicVersion: GetVersion(),
			CompilerVersion:   plugin.Request.CompilerVersion.String(),
			File:              file.GoImportPath.String(),
			UseMeta:           IfNill(apiGateway.UseMeta, false),
			UseValidation:     IfNill(apiGateway.UseValidation, false),
		}
		path := strings.Split(strings.ReplaceAll(file.GeneratedFilenamePrefix, "\\", "/"), "/")
		filename := moduleName + "/" + strings.Join(path[:len(path)-1], "/") + "/" + fmt.Sprintf("%s.gateway.pb.go", service.GoName)
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
