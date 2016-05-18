// Copyright 2016 Alex Fluter

package lang

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/scanner"
	"go/token"
	"log"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/tools/imports"
)

type Go struct {
	path  string
	fsrc  string
	fprog string
	opt   *imports.Options
}

func (g *Go) Name() string {
	return "Go"
}

func (g *Go) Version() string {
	var stdout bytes.Buffer
	err := runLocal("go", []string{"version"}, ".", nil, &stdout, nil)
	if err != nil {
		return "Unknown"
	}
	return stdout.String()
}

func (g *Go) Init() error {
	path, err := exec.LookPath("go")
	if err != nil {
		return err
	}
	g.path = path
	g.fsrc = "source.go"
	g.fprog = "prog.go"
	g.opt = &imports.Options{
		Fragment: true,
	}
	return nil
}

func (g *Go) Compile(code string) *Result {
	var result Result
	var err error
	var stdout, stderr bytes.Buffer
	var args []string
	var dir string
	var filetorun string

	dir, err = setupWorkspace(g, g.fsrc, code)
	if err != nil {
		return &Result{Error: err.Error()}
	}
	filetorun = g.fsrc

	var source string = code
	var fset *token.FileSet
	fset = token.NewFileSet()
	_, err = parser.ParseFile(fset, "stdin", code, 0)
	if err != nil {
		if el, ok := err.(scanner.ErrorList); ok {
			if strings.HasPrefix(el[0].Msg,
				"expected 'package', found 'IDENT'") {
				source = fmt.Sprintf("func main() {\n%s\n}", code)
			}
		} else {
			result.Error = err.Error()
			return &result
		}
	}

	processed, err := imports.Process("stdin", []byte(source), g.opt)
	if err != nil {
		result.Error = err.Error()
		return &result
	}

	err = writeSource(fmt.Sprintf("%s/%s", dir, g.fprog),
		string(processed))
	if err != nil {
		result.Error = err.Error()
		return &result
	}
	filetorun = g.fprog

	args = []string{"run", filetorun}
	err = runTimed(g.path,
		args,
		dir,
		nil,
		&stdout,
		&stderr,
		RunTimeout*time.Second)
	if err != nil {
		log.Println(err)
		result.Error = err.Error()
	}
	result.Cmd = strings.Join(append([]string{"go"}, args...), " ")
	result.P_Output = getStringBuffer(&stdout)
	result.P_Error = getStringBuffer(&stderr)
	return &result
}
