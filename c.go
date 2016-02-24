// Copyright 2016 Alex Fluter

package lotsawa

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

type CCompiler struct {
	path    string
	options []string
}

func (c *CCompiler) Name() string {
	return "GCC"
}

func (c *CCompiler) Init() error {
	var path string
	var err error

	path, err = exec.LookPath("gcc")
	if err != nil {
		return err
	}
	log.Println("gcc is:", path)
	c.path = path

	c.options = []string{
		"-Wextra",
		"-Wall",
		"-Wno-unused",
		"-pedantic",
		"-Wfloat-equal",
		"-Wshadow",
		"-std=c11",
		"-lm",
		"-Wfatal-errors",
		"-fsanitize=alignment,undefined"}
	return nil
}

func (c *CCompiler) Version() string {
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

func (c *CCompiler) Compile(code string) *Result {
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
		log.Println("Failed to create data store:", err)
		result.err = err.Error()
		return &result
	}

	srcReader = bytes.NewReader([]byte(code))

	srcFile, objFile, execFile =
		fmt.Sprintf("%s/%s", dir, Fsrc),
		fmt.Sprintf("%s/%s", dir, Fobj),
		fmt.Sprintf("%s/%s", dir, Fbin)

	fsrc, err = os.Create(srcFile)
	if err != nil {
		result.err = err.Error()
		return &result
	}
	_, err = io.Copy(fsrc, srcReader)

	result.main = c.detectMain(code)

	srcReader.Seek(0, 0)
	if !result.main {
		args = append(c.options, "-xc", "-o", objFile, "-c", "-")

		err = run(c.path, args, "", srcReader, &stdOut, &stdErr)
		result.cmd = strings.Join(args, " ")
		result.c_out, result.c_err = stdOut.String(), stdErr.String()
		if err != nil {
			result.err = "gcc: " + err.Error()
			return &result
		}
	} else {
		args = append(c.options, "-xc", "-o", execFile, "-")

		err = run(c.path, args, "", srcReader, &stdOut, &stdErr)
		result.cmd = strings.Join(args, " ")
		result.c_out, result.c_err = stdOut.String(), stdErr.String()
		if err != nil {
			result.err = "gcc: " + err.Error()
			return &result
		}

		var execOut, execErr bytes.Buffer
		err = runTimed(execFile, nil, "", nil, &execOut, &execErr, runTimeout*time.Second)
		if err != nil {
			log.Println("error run:", err)
			result.err = execFile + ": " + err.Error()
		}

		if execOut.Len() < 500 {
			result.p_out = execOut.String()
		} else {
			result.p_out = string(execOut.Next(500)) + "..."
		}
		if execErr.Len() < 500 {
			result.p_err = execErr.String()
		} else {
			result.p_err = string(execOut.Next(500)) + "..."
		}
	}

	return &result
}

func (c *CCompiler) detectMain(code string) bool {
	if mainRe.FindString(code) != "" {
		return true
	}
	return false
}
