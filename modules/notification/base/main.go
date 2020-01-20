// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
	"text/template"
	"time"
)

type funcDef struct {
	Name string
	Args []funcDefArg
}

type funcDefArg struct {
	Name string
	Type string
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "notifier.go", nil, 0)
	if err != nil {
		panic(err)
	}
	funcs := make([]funcDef, 0)
	//currentFunc := funcDef{}
	ast.Inspect(f, func(n ast.Node) bool {
		spec, ok := n.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "Notifier" {
			return true
		}
		child, ok := spec.Type.(*ast.InterfaceType)
		if !ok {
			return false
		}
		funcs = make([]funcDef, len(child.Methods.List))
		for i, method := range child.Methods.List {
			methodFuncDef := method.Type.(*ast.FuncType)
			def := funcDef{}
			def.Name = method.Names[0].Name
			def.Args = make([]funcDefArg, 0, len(methodFuncDef.Params.List))
			for j, param := range methodFuncDef.Params.List {
				defaultName := fmt.Sprintf("unknown%d", j)
				sb := strings.Builder{}
				format.Node(&sb, fset, param.Type)

				if len(param.Names) == 0 {
					def.Args = append(def.Args, funcDefArg{
						Name: defaultName,
						Type: sb.String(),
					})
				} else {
					for _, ident := range param.Names {
						def.Args = append(def.Args, funcDefArg{
							Name: ident.Name,
							Type: sb.String(),
						})
					}
				}
			}
			funcs[i] = def
		}

		return true
	})

	buf := bytes.Buffer{}
	nullTemplate.Execute(&buf, struct {
		Timestamp time.Time
		Funcs     []funcDef
	}{
		Timestamp: time.Now(),
		Funcs:     funcs,
	})

	bs, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("null.go", bs, 0644)
	if err != nil {
		panic(err)
	}

	buf = bytes.Buffer{}
	queueTemplate.Execute(&buf, struct {
		Timestamp time.Time
		Funcs     []funcDef
	}{
		Timestamp: time.Now(),
		Funcs:     funcs,
	})

	bs, err = format.Source(buf.Bytes())
	if err != nil {
		ioutil.WriteFile("queue.go", buf.Bytes(), 0644)
		panic(err)
	}

	err = ioutil.WriteFile("queue.go", bs, 0644)
	if err != nil {
		panic(err)
	}

}

var queueTemplate = template.Must(template.New("").Parse(`
// Code generated by go generate; DO NOT EDIT.
package base

import (
	"encoding/json"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/graceful"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/repository"
	"code.gitea.io/gitea/modules/queue"
)

// FunctionCall represents is function call with json.Marshaled arguments
type FunctionCall struct {
	Name string
	Args [][]byte
}

type QueueNotifier struct {
	name string
	notifiers []Notifier
	internal queue.Queue
}

var (
	_ Notifier = &QueueNotifier{}
)

func NewQueueNotifier(name string, notifiers []Notifier) Notifier {
	q := &QueueNotifier{
		name: name,
		notifiers: notifiers,
	}
	q.internal = queue.CreateQueue(name, q.handle, &FunctionCall{})
	return q
}

func NewQueueNotifierWithHandle(name string, handle queue.HandlerFunc) Notifier {
	q := &QueueNotifier{
		name: name,
	}
	q.internal = queue.CreateQueue(name, handle, &FunctionCall{})
	return q
}

func (q *QueueNotifier) handle(data ...queue.Data) {
	for _, datum := range data {
		call := datum.(*FunctionCall)
		var err error
		switch call.Name {
		{{- range .Funcs }}
		case "{{.Name}}":
			{{$p := .Name}}
			{{- range $i, $e := .Args }}
			var {{$e.Name}} {{$e.Type}}
			err = json.Unmarshal(call.Args[{{$i}}], &{{$e.Name}})
			if err != nil {
				log.Error("Unable to unmarshal %s to %s in call to %s: %v", string(call.Args[{{$i}}]), "{{$e.Type}}", "{{$p}}", err)
				continue
			}
			{{- end }}
			for _, notifier := range q.notifiers {
				notifier.{{.Name}}({{- range $i, $e := .Args}}{{ if $i }}, {{ end }}{{$e.Name}}{{end}})
			}
		{{- end }}
		default:
			log.Error("Unknown notifier function %s with %d arguments", call.Name, len(call.Args))
		}
	}
}

func (q *QueueNotifier) Run() {
	for _, notifier := range q.notifiers {
		go notifier.Run()
	}
	graceful.GetManager().RunWithShutdownFns(q.internal.Run)
}
{{- range .Funcs}}
{{if ne .Name "Run"}}

// {{ .Name }} is a placeholder function
func (q *QueueNotifier) {{ .Name }}({{ range $i, $e := .Args }}{{ if $i }}, {{ end }}{{$e.Name}} {{$e.Type}}{{end}}) {
	args := make([][]byte, 0)
	var err error
	var bs []byte
	{{- range .Args }}
	bs, err = json.Marshal(&{{.Name}})
	if err != nil {
		log.Error("Unable to marshall {{.Name}}: %v", err)
		return
	}
	args = append(args, bs)
	{{- end }}

	q.internal.Push(&FunctionCall{
		Name: "{{.Name}}",
		Args: args,
	})
}
{{end}}
{{- end }}
`))

var nullTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
package base

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/repository"
)

// NullNotifier implements a blank notifier
type NullNotifier struct {
}

var (
	_ Notifier = &NullNotifier{}
)
{{- range .Funcs}}

// {{ .Name }} is a placeholder function
func (*NullNotifier) {{ .Name }}({{ range $i, $e := .Args }}{{ if $i }}, {{ end }}{{$e.Name}} {{$e.Type}}{{end}}) {}
{{- end }}
`))
