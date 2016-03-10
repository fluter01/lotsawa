// Copyright 2016 Alex Fluter

package lang

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os/exec"
	"strings"
	"time"
)

var goBuiltinTypes []string = []string{
	"ComplexType",
	"FloatType",
	"IntegerType",
	"Type",
	"Type1",
	"bool",
	"byte",
	"complex128",
	"complex64",
	"error",
	"float32",
	"float64",
	"int",
	"int16",
	"int32",
	"int64",
	"int8",
	"rune",
	"string",
	"uint",
	"uint16",
	"uint32",
	"uint64",
	"uint8",
	"uintptr",
}

type Go struct {
	path         string
	fsrc         string
	fannotated   string
	builtinTypes map[string]struct{}
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
	g.fsrc = "prog.go"
	g.fannotated = "source.go"
	g.builtinTypes = make(map[string]struct{})
	for _, t := range goBuiltinTypes {
		g.builtinTypes[t] = struct{}{}
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

	var file *ast.File
	var fset *token.FileSet
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, "stdin", code, 0)
	if err != nil {
		// add package and func
		if file.Name.Name == "" {
			var pkgs map[string]bool = make(map[string]bool)
			var imports string
			source := fmt.Sprintf(`package main

func main() {
	%s
}`, code)
			file, err = parser.ParseFile(fset, "stdin", source, 0)
			for _, ident := range file.Unresolved {
				pkgs[ident.Name] = true
			}
			for k := range pkgs {
				if _, ok := g.builtinTypes[k]; ok {
					continue
				}
				imports += fmt.Sprintf("import \"%s\"\n", k)
			}
			source = fmt.Sprintf(`package main
%s
func main() {
	%s
}`, imports, code)
			err = writeSource(fmt.Sprintf("%s/%s", dir, g.fannotated), source)
			if err != nil {
				result.Error = err.Error()
				return &result
			}
			filetorun = g.fannotated
		} else {
			result.Error = err.Error()
			return &result
		}
	}

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
