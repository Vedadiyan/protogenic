package protogenic

import (
	"bytes"
	"fmt"
	"strings"
)

func StringToGoByteArray(str string) string {
	buffer := bytes.NewBufferString("[]byte {")
	line := bytes.NewBufferString("")
	for i := 0; i < len(str); i++ {
		if i != 0 && i%16 == 0 {
			buffer.WriteString("\r\n")
			buffer.WriteString("\t")
			buffer.Write(line.Bytes())
			line.Reset()
		}
		line.WriteString(fmt.Sprintf("0x%02x, ", str[i]))
	}
	buffer.WriteString("\r\n")
	buffer.WriteString("\t")
	buffer.Write(line.Bytes())
	buffer.WriteString("\r\n")
	buffer.WriteString("}")
	return buffer.String()
}

func EmptyIfNill(str *string) string {
	if str != nil {
		return *str
	}
	return ""
}

func CombinePath(path ...string) string {
	buffer := bytes.NewBufferString("")
	for i := 0; i < len(path); i++ {
		buffer.WriteString(strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(path[i], "/"), "/"), "\\", "/"))
		buffer.WriteString("/")
	}
	return buffer.String()
}
