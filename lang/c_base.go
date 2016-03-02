// Copyright 2016 Alex Fluter

package lang

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	runTimeout = 3
)

const (
	Fsrc = "prog.c"
	Fobj = "prog.o"
	Fbin = "prog"
)

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

func (c *CBase) compile(code, prelude string) *Result {
	var err error
	var fsrc *os.File
	var srcReader *bytes.Reader
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	var result Result
	var dir string
	var srcFile, objFile, execFile string
	var args []string

	dir, err = ioutil.TempDir(DataStore, c.Name())
	if err != nil {
		log.Println("Failed to create workspace:", err)
		result.Error = err.Error()
		return &result
	}

	err = os.Chmod(dir, 0775)
	if err != nil {
		log.Println("Failed to chown:", err)
		result.Error = err.Error()
		return &result
	}

	srcReader = bytes.NewReader([]byte(prelude + code))

	srcFile, objFile, execFile =
		fmt.Sprintf("%s/%s", dir, Fsrc),
		fmt.Sprintf("./%s", Fobj),
		fmt.Sprintf("./%s", Fbin)

	fsrc, err = os.Create(srcFile)
	if err != nil {
		result.Error = err.Error()
		return &result
	}
	_, err = io.Copy(fsrc, srcReader)

	main := c.detectMain(code)

	srcReader.Seek(0, 0)
	if !main {
		args = append(c.options, "-xc", "-o", objFile, "-c", "-")

		err = run(c.path, args, dir, srcReader, &stdOut, &stdErr)
		result.Cmd = strings.Join(args, " ")
		result.C_Output, result.C_Error = stdOut.String(), stdErr.String()
		if err != nil {
			result.Error = "gcc: " + err.Error()
			return &result
		}
	} else {
		args = append(c.options, "-xc", "-o", execFile, "-")

		err = run(c.path, args, dir, srcReader, &stdOut, &stdErr)
		result.Cmd = strings.Join(args, " ")
		result.C_Output, result.C_Error = stdOut.String(), stdErr.String()
		if err != nil {
			result.Error = "gcc: " + err.Error()
			return &result
		}

		var execOut, execErr bytes.Buffer

		if use_container {
			err = runContainerTimed(execFile, nil, dir, nil,
				&execOut, &execErr, runTimeout*time.Second)
		} else {
			err = runTimed(execFile, nil, dir, nil,
				&execOut, &execErr, runTimeout*time.Second)
		}
		if err != nil {
			log.Println("error run:", err)
			result.Error = execFile + ": " + err.Error()
		}

		if execOut.Len() < 500 {
			result.P_Output = execOut.String()
		} else {
			result.P_Output = string(execOut.Next(500)) + "..."
		}
		if execErr.Len() < 500 {
			result.P_Error = execErr.String()
		} else {
			result.P_Error = string(execOut.Next(500)) + "..."
		}
	}

	return &result
}

func (c *CBase) detectMain(code string) bool {
	if mainRe.FindString(code) != "" {
		return true
	}
	return false
}
