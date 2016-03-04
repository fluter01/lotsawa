// Copyright 2016 Alex Fluter

package lang

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

type Bash struct {
	path string
	fsrc string
}

func (sh *Bash) Name() string {
	return "Bash"
}

func (sh *Bash) Version() string {
	res := sh.Compile("echo $BASH_VERSION")
	if res.Error != "" {
		return "Unknown"
	}
	return res.P_Output
}

func (sh *Bash) Init() error {
	var err error
	var path string

	path, err = exec.LookPath("bash")
	if err != nil {
		return err
	}
	sh.path = path
	sh.fsrc = "prog.sh"

	return nil
}

func (sh *Bash) Compile(code string) *Result {
	var result Result
	var err error
	var stdout, stderr bytes.Buffer
	var args []string
	var dir string

	dir, err = createWorkspace(sh)
	if err != nil {
		return &Result{Error: err.Error()}
	}

	srcpath := fmt.Sprintf("%s/%s", dir, sh.fsrc)
	err = writeSource(srcpath, code)
	if err != nil {
		return &Result{Error: err.Error()}
	}

	args = []string{"-c", code}
	err = runTimed(sh.path,
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
