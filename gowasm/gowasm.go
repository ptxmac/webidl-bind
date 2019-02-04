// Go WASM output
package gowasm

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const fileTemplInput = `
{{define "header"}}
package {{.Package}}

import "syscall/js"

// ReleasableApiResource is used to release underlaying
// allocated resources.
type ReleasableApiResource interface {
	Release()
}

type releasableApiResourceList []ReleasableApiResource

func (a releasableApiResourceList) Release() {
	for _, v := range a {
		v.Release()
	}
}

{{end}}
`

var fileTempl = template.Must(template.New("file").Parse(fileTemplInput))

type fileData struct {
	Package string
}

type writeFn func(dst io.Writer, in types.Type) error

// WriteSource is create source code files.
// returns map["path/filename"]"file content"
func WriteSource(conv *types.Convert) (map[string][]byte, error) {
	target := make(map[string]*bytes.Buffer)
	var err error
	for _, e := range conv.Enums {
		if e.InUse() {
			err = writeType(e, target, writeEnum, err)
		}
	}
	for _, v := range conv.Callbacks {
		if v.InUse() {
		}
		err = writeType(v, target, writeCallback, err)
	}
	for _, v := range conv.Dictionary {
		if v.InUse() {
			err = writeType(v, target, writeDictionary, err)
		}
	}
	for _, v := range conv.Interface {
		if v.InUse() {
			err = writeType(v, target, writeInterface, err)
		}
	}
	if err != nil {
		return nil, err
	}
	ret := make(map[string][]byte)
	for k, v := range target {
		low := strings.ToLower(k)
		filename := fmt.Sprintf("%s/%s.go", low, low)
		content := v.Bytes()
		content = sourceCodeRemoveEmptyLines(content)
		if source, err := format.Source(content); err == nil {
			content = source
		} else {
			// we just print this error to get an output file that we
			// later can correct and fix the bug
			fmt.Fprintf(os.Stderr, "error:%s:unable to format output source code: %s\n", filename, err)
		}
		ret[filename] = content
	}
	return ret, nil
}

func writeType(value types.Type, target map[string]*bytes.Buffer, conv writeFn, err error) error {
	if err != nil {
		return err
	}
	dst, err := getTarget(value, target)
	if err != nil {
		return err
	}
	if err := conv(dst, value); err != nil {
		return err
	}
	return nil
}

func getTarget(value types.Type, target map[string]*bytes.Buffer) (*bytes.Buffer, error) {
	pkg := value.Basic().Package
	dst, ok := target[pkg]
	if ok {
		return dst, nil
	}
	dst = &bytes.Buffer{}
	target[pkg] = dst
	data := fileData{
		Package: pkg,
	}
	if err := fileTempl.ExecuteTemplate(dst, "header", data); err != nil {
		return nil, err
	}
	return dst, nil
}

func FormatPkg(filename string) string {
	value := filepath.Base(filename)
	idx := strings.Index(value, ".")
	if idx != -1 {
		return value[0:idx]
	}
	value = strings.ToLower(value)
	return value
}

// sourceCodeRemoveEmptyLines will remove empty lines
func sourceCodeRemoveEmptyLines(code []byte) []byte {
	add := []string{"//", "func", "type", "const", "var"}
	in := bytes.NewBuffer(code)
	var out bytes.Buffer
	ignore := false
	for {
		s, err := in.ReadString('\n')
		if err != nil && err != io.EOF {
			panic(err)
		}
		if err == io.EOF {
			break
		}
		if len(strings.TrimSpace(s)) == 0 {
			continue
		}
		found := false
		for _, prefix := range add {
			if strings.HasPrefix(s, prefix) {
				found = true
				if !ignore {
					out.WriteByte('\n')
				}
				ignore = true
			}
		}
		if !found {
			ignore = false
		}
		out.WriteString(s)
	}
	return out.Bytes()
}
