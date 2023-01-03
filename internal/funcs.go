package protogenic

import (
	"bytes"
	"strings"
	"text/template"
)

const (
	_indent = "{{$INDENT$}}"
)

var (
	_funcs template.FuncMap
)

func init() {
	_funcs = template.FuncMap{
		"ToLowerCase": strings.ToLower,
		"ToUpperCase": strings.ToUpper,
		"ToCamelCase": func(str string) string {
			return changeCase(str, false)
		},
		"ToPascalCase": func(str string) string {
			return changeCase(str, true)
		},
		"Indent": func(indent int) string {
			str := bytes.NewBufferString(_indent)
			for i := 0; i < indent; i++ {
				str.WriteString(" ")
			}
			return str.String()
		},
	}
}

func PostProcess(strBuffer bytes.Buffer) bytes.Buffer {
	buffer := bytes.NewBufferString("")
	lines := strings.Split(strBuffer.String(), "\n")
	for _, line := range lines {
		tmp := strings.TrimLeftFunc(line, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\r'
		})
		if len(tmp) == 0 {
			continue
		}
		if !strings.HasPrefix(tmp, _indent) {
			buffer.WriteString(line)
			continue
		}
		tmp = strings.TrimPrefix(tmp, _indent)
		buffer.WriteString(tmp)
	}
	return *buffer
}
