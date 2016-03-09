// Copyright 2016 Alex Fluter

package lang

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
	"time"
)

type Go struct {
	path string
	fsrc string
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
	return nil
}

func (g *Go) Compile(code string) *Result {
	var result Result
	var err error
	var stdout, stderr bytes.Buffer
	var args []string
	var dir string

	dir, err = setupWorkspace(g, g.fsrc, code)
	if err != nil {
		return &Result{Error: err.Error()}
	}

	args = []string{"run", g.fsrc}
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
		return &result
	}
	result.Cmd = strings.Join(args, " ")
	result.P_Output = getStringBuffer(&stdout)
	result.P_Error = getStringBuffer(&stderr)
	return &result
}
