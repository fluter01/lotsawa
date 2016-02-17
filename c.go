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
)

const (
	DataStore = "store"
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
	var runCmd *exec.Cmd
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

		runCmd = exec.Command(c.path, args...)
		runCmd.Stdout = &stdOut
		runCmd.Stderr = &stdErr
		runCmd.Stdin = srcReader
		err = runCmd.Run()
		result.cmd = strings.Join(runCmd.Args, " ")
		result.c_out, result.c_err = stdOut.String(), stdErr.String()
		if err != nil {
			result.err = "gcc: " + err.Error()
			return &result
		}
	} else {
		args = append(c.options, "-xc", "-o", execFile, "-")
		runCmd = exec.Command(c.path, args...)
		runCmd.Stdout = &stdOut
		runCmd.Stderr = &stdErr
		runCmd.Stdin = srcReader
		err = runCmd.Run()
		result.cmd = strings.Join(runCmd.Args, " ")
		result.c_out, result.c_err = stdOut.String(), stdErr.String()
		if err != nil {
			result.err = "gcc: " + err.Error()
			return &result
		}

		var execOut, execErr bytes.Buffer
		runCmd = exec.Command(execFile)
		runCmd.Stdout = &execOut
		runCmd.Stderr = &execErr
		err = runCmd.Run()
		result.p_out, result.p_err = execOut.String(), execErr.String()
		if err != nil {
			log.Println("error run:", err)
			result.err = execFile + ": " + err.Error()
			return &result
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
