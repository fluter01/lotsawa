// Copyright 2016 Alex Fluter

package lang

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const ()

// int main()
// int main(void)
// int main(int argc, char* argv[])
// int main(int argc, char **argv)
// void main()
// void main(void)
// void main(int argc, char* argv[])
// void main(int argc, char **argv)

const mainPtn = "(int|void)\\s+main\\s*\\("

var mainRe = regexp.MustCompile(mainPtn)

// The base compiler for C language
type CBase struct {
	path    string
	prelude string
	options []string
	fsrc    string
	fobj    string
	fbin    string
}

func (c *CBase) Name() string {
	return "GCC Base"
}

func (c *CBase) Init() error {
	var path string
	var err error

	path, err = exec.LookPath("gcc")
	if err != nil {
		return err
	}
	c.path = path

	c.options = []string{}
	c.prelude = ""
	c.fsrc = "prog.c"
	c.fobj = "prog.o"
	c.fbin = "prog"
	return nil
}

func (c *CBase) Version() string {
	var runCmd *exec.Cmd
	var err error
	var stdOut bytes.Buffer

	runCmd = exec.Command(c.path, "--version")
	runCmd.Stdout = &stdOut
	err = runCmd.Run()
	if err != nil {
		log.Println("error exec ", c.path, ":", err)
		return ""
	}
	return stdOut.String()
}

func (c *CBase) compile(caller Compiler, code, prelude string) *Result {
	var err error
	var srcReader *bytes.Reader
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	var result Result
	var dir string
	var srcFile, objFile, execFile string
	var args []string

	if caller == nil {
		return nil
	}
	dir, err = createWorkspace(caller)
	if err != nil {
		log.Println("Failed to setup workspace:", err)
		result.Error = err.Error()
		return &result
	}

	srcReader = bytes.NewReader([]byte(prelude + code))

	srcFile, objFile, execFile =
		fmt.Sprintf("%s/%s", dir, c.fsrc),
		fmt.Sprintf("./%s", c.fobj),
		fmt.Sprintf("./%s", c.fbin)

	err = writeSource(srcFile, code)
	if err != nil {
		log.Println("Failed to write source:", err)
		result.Error = err.Error()
		return &result
	}
	main := c.detectMain(code)

	if !main {
		args = append(c.options, "-xc", "-o", objFile, "-c", "-")

		err = runLocal(c.path, args, dir, srcReader, &stdOut, &stdErr)
		result.Cmd = strings.Join(args, " ")
		result.C_Output, result.C_Error =
			getStringBuffer(&stdOut), getStringBuffer(&stdErr)
		if err != nil {
			result.Error = "gcc: " + err.Error()
			return &result
		}
	} else {
		args = append(c.options, "-xc", "-o", execFile, "-")

		err = runLocal(c.path, args, dir, srcReader, &stdOut, &stdErr)
		result.Cmd = strings.Join(args, " ")
		result.C_Output, result.C_Error =
			getStringBuffer(&stdOut), getStringBuffer(&stdErr)
		if err != nil {
			result.Error = "gcc: " + err.Error()
			return &result
		}

		var execOut, execErr bytes.Buffer

		err = runTimed(execFile, nil, dir, nil,
			&execOut, &execErr, RunTimeout*time.Second)
		if err != nil {
			log.Println("error run:", err)
			result.Error = execFile + ": " + err.Error()
		}

		result.P_Output, result.P_Error =
			getStringBuffer(&execOut), getStringBuffer(&execErr)
	}

	return &result
}

func (c *CBase) detectMain(code string) bool {
	if mainRe.FindString(code) != "" {
		return true
	}
	return false
}
