// +build tools

package main

import (
	"bytes"
	"go/format"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

func main() {
	// IMPORTANT: The name of this file must sort lexigraphically before main.go.
	// Initialization order is as follows:
	//   * variable initialization (via dependency graph)
	//   * init() in each file (in lexical filename sort order)
	//   * main() as is defined only once
	file := "about.go"

	version := getVersion(os.Getenv("GITEA_VERSION"), os.Getenv("DRONE_TAG"), os.Getenv("VERSION"))
	tags := os.Getenv("TAGS")

	// Set all version info for templates
	v := meta{
		Version: version,
		Tags:    tags,
	}

	// Create or overwrite the go file from template
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, v); nil != err {
		panic(err)
	}
	src, err := format.Source(buf.Bytes())
	if nil != err {
		panic(err)
	}
	f, err := os.Create(file)
	if nil != err {
		panic(err)
	}
	defer f.Close()
	f.Write(src)
}

type meta struct {
	Tags    string
	Version string
}

func getVersion(gitea, drone, verenv string) string {
	// GITEA_VERSION takes priority, always
	if 0 != len(gitea) {
		return gitea
	}

	// DRONE_TAG version comes next
	if 0 != len(drone) {
		// Only when DRONE_TAG is set does VERSION takes precedence
		// over both DRONE_TAG and the GIT DESCRIPTION
		if 0 != len(verenv) {
			return verenv
		}
		return strings.TrimLeft(drone, "v")
	}

	// Usually, however, we just shell out and get the version from git
	tagcmd := strings.Split("git describe --tags --always", " ")
	cmd := exec.Command(tagcmd[0], tagcmd[1:]...)
	out, err := cmd.CombinedOutput()
	if nil != err {
		panic(strings.Join(tagcmd, " ") + ": " + err.Error() + "\n" + string(out))
	}
	desc := strings.TrimSpace(string(out))
	return strings.TrimLeft(strings.Replace(desc, "-", "+", -1), "v")
}

var tpl = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
package main

func init() {
	Version = "{{ .Version }}"
	Tags = "{{ .Tags }}"
	MakeVersion = ""
}
`))
