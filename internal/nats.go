package protogenic

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"

	_ "embed"
)

var (
	//go:embed templates/nats/client.go.tmpl
	_client string
	//go:embed templates/nats/server.go.tmpl
	_server string
	//go:embed templates/nats/common.go.tmpl
	_common string
)

type Nats struct {
	Name      string
	Request   string
	Response  string
	Logger    bool
	Namespace string
	Queue     *string
}

type NatsContext struct {
	NatsServices []Nats
	Package      string
}

func GenerateNats(plugin *protogen.Plugin, file *protogen.File) error {
	serverTemplate, err := template.New("server").Funcs(_funcs).Parse(_server)
	if err != nil {
		return err
	}
	clientTemplate, err := template.New("client").Funcs(_funcs).Parse(_client)
	if err != nil {
		return err
	}
	commonTemplate, err := template.New("server").Funcs(_funcs).Parse(_common)
	if err != nil {
		return err
	}
	natsServices := make([]Nats, 0)
	for _, service := range file.Services {
		filename := file.GeneratedFilenamePrefix + ".%s.nats" + ".pb.go"
		server := plugin.NewGeneratedFile(strings.ToLower(fmt.Sprintf(filename, "server")), file.GoImportPath)
		client := plugin.NewGeneratedFile(strings.ToLower(fmt.Sprintf(filename, "client")), file.GoImportPath)
		common := plugin.NewGeneratedFile(strings.ToLower(fmt.Sprintf(filename, "common")), file.GoImportPath)
		for _, method := range service.Methods {
			namespace, queue := getNamespace(method.Comments.Leading)
			natsService := Nats{
				Name:      method.GoName,
				Request:   method.Input.GoIdent.GoName,
				Response:  method.Output.GoIdent.GoName,
				Logger:    true,
				Namespace: namespace,
				Queue:     &queue,
			}
			natsServices = append(natsServices, natsService)
		}
		natsContext := NatsContext{
			NatsServices: natsServices,
			Package:      string(file.GoPackageName),
		}
		var serverCode, clientCode, commonCode bytes.Buffer
		err := serverTemplate.Execute(&serverCode, natsContext)
		if err != nil {
			return err
		}
		err = clientTemplate.Execute(&clientCode, natsContext)
		if err != nil {
			return err
		}
		err = commonTemplate.Execute(&commonCode, natsContext)
		if err != nil {
			return err
		}
		server.P(serverCode.String())
		client.P(clientCode.String())
		common.P(commonCode.String())
	}
	return nil
}

func getNamespace(comments protogen.Comments) (namespace string, queue string) {
	var ns string = ""
	var q string = ""
	for _, comment := range strings.Split(comments.String(), "\n") {
		tmp := AlignLeft(comment)
		if strings.HasPrefix(tmp, "@namespace") {
			ns = ExtractString(tmp)
			continue
		}
		if strings.HasPrefix(tmp, "@queue") {
			q = ExtractString(tmp)
			continue
		}
	}
	return ns, q
}

var characters = []string{
	" ",
	"/",
	"\t",
	"\r",
	"\n",
}

func ExtractString(value string) string {
	segments := strings.Split(value, " ")
	if len(segments) >= 2 {
		val := strings.TrimRight(strings.Join(segments[1:], " "), "\r\n")
		return AlignRight(val)
	}
	return ""
}

func AlignLeft(str string) string {
	output := str
	for {
		_str := output
		for _, value := range characters {
			_str = strings.TrimLeft(_str, value)
		}
		if _str == output {
			break
		}
		output = _str
	}
	return output
}

func AlignRight(str string) string {
	output := str
	for {
		_str := output
		for _, value := range characters {
			_str = strings.TrimRight(_str, value)
		}
		if _str == output {
			break
		}
		output = _str
	}
	return output
}
