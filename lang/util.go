// Copyright 2016 Alex Fluter

package lang

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

func run(name string,
	args []string,
	wd string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer) error {
	if use_container {
		return runContainer(name, args, wd, stdin, stdout, stderr)
	}
	return runLocal(name, args, wd, stdin, stdout, stderr)
}

func runTimed(name string,
	args []string,
	wd string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	timeout time.Duration) error {
	if use_container {
		return runContainerTimed(name, args, wd, stdin, stdout, stderr, timeout)
	}
	return runLocalTimed(name, args, wd, stdin, stdout, stderr, timeout)
}

func runLocal(name string,
	args []string,
	wd string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer) error {
	var err error
	var cmd *exec.Cmd

	cmd = exec.Command(name, args...)
	cmd.Dir = wd
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()

	if err != nil {
		return err
	}
	return nil
}

func runLocalTimed(name string,
	args []string,
	wd string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	timeout time.Duration) error {
	var err error
	var cmd *exec.Cmd

	cmd = exec.Command(name, args...)
	cmd.Dir = wd
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	chDone := make(chan bool)
	go func() {
		err = cmd.Run()
		chDone <- true
	}()

	var done bool = false
	select {
	case done = <-chDone:
		break
	case t := <-time.After(timeout):
		if !done {
			cmd.Process.Kill()
			var st *os.ProcessState
			st, _ = cmd.Process.Wait()
			err = fmt.Errorf("program killed after %s:%s",
				t.Sub(start).String(), st)
		}
		break
	}

	return err
}

func createWorkspace(c Compiler) (string, string, error) {
	var dir string
	var err error

	dir, err = ioutil.TempDir(DataStore, c.Name())
	if err != nil {
		return "", "", err
	}
	prefix := path.Clean(fmt.Sprintf("%s/%s", DataStore, c.Name()))
	err = os.Chmod(dir, 0775)
	if err != nil {
		return "", "", err
	}

	id := dir[len(prefix):]
	return dir, id, nil
}

func writeSource(path, code string) error {
	var err error
	var src *bytes.Reader

	src = bytes.NewReader([]byte(code))
	fsrc, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fsrc.Close()
	_, err = io.Copy(fsrc, src)
	return err
}

func setupWorkspace(c Compiler, sourcefile, code string) (string, string, error) {
	dir, id, err := createWorkspace(c)
	if err != nil {
		return "", "", err
	}
	srcpath := fmt.Sprintf("%s/%s", dir, sourcefile)
	err = writeSource(srcpath, code)
	if err != nil {
		return "", "", err
	}
	return dir, id, nil
}

func getStringBuffer(buf *bytes.Buffer) string {
	if buf == nil {
		return ""
	}
	if buf.Len() <= MaxLength {
		return buf.String()
	}

	return string(buf.Next(256)) + "..."
}
