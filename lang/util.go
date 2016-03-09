// Copyright 2016 Alex Fluter

package lang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/specs"
)

const (
	fconf = "libcontainer.json"
)

var (
	master_config *configs.Config
	factory       libcontainer.Factory
	use_container bool
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

func init() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
		factory, _ := libcontainer.New("")
		if err := factory.StartInitialization(); err != nil {
			log.Fatal(err)
		}
		panic("--this line should have never been executed, congratulations--")
	}
}

func InitContainer() error {
	var err error

	factory, err = createFactory()
	if err != nil {
		return err
	}

	master_config, err = loadConfig(fconf)
	if err != nil {
		return err
	}
	use_container = true
	return nil
}

func loadConfig(path string) (*configs.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file %s not found", path)
		}
		return nil, err
	}
	defer f.Close()

	var config configs.Config
	if err = json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func createFactory() (libcontainer.Factory, error) {
	root := specs.LinuxStateDirectory
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return libcontainer.New(abs, libcontainer.Cgroupfs, func(l *libcontainer.LinuxFactory) error {
		l.CriuPath = "criu"
		return nil
	})
}

func createProcess(name string, stdin io.Reader, stdout, stderr io.Writer) *libcontainer.Process {
	return &libcontainer.Process{
		Args:   []string{name},
		Env:    []string{"PATH=/bin:/sbin:/usr/bin:/usr/sbin"},
		User:   "root",
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
}
func runContainer(name string,
	args []string,
	wd string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer) error {
	var err error
	var id string

	id = path.Base(wd)

	// mount base rootfs with working directory
	rootfs := master_config.Rootfs
	lowerdir := rootfs
	upperdir, err := filepath.Abs(wd)
	if err != nil {
		return err
	}
	workdir, err := filepath.Abs(fmt.Sprintf("%s-%s", wd, "work"))
	if err != nil {
		return err
	}

	err = os.Mkdir(workdir, 0775)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer os.RemoveAll(workdir)
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		lowerdir, upperdir, workdir)
	err = syscall.Mount("overlay", upperdir, "overlay", syscall.MS_MGC_VAL,
		opts)
	if err != nil {
		return err
	}
	defer func() {
		err := syscall.Unmount(upperdir, 0)
		if err != nil {
			return
		}
	}()

	// set cgroup path
	var config configs.Config
	config = *master_config
	config.Cgroups.Path = fmt.Sprintf("%s/%s",
		config.Cgroups.Path, id)
	config.Rootfs = upperdir
	container, err := factory.Create(id, &config)
	if err != nil {
		return err
	}
	defer container.Destroy()

	args = append([]string{name}, args...)
	process := &libcontainer.Process{
		Args:   args,
		Env:    []string{"PATH=/bin:/sbin:/usr/bin:/usr/sbin"},
		User:   "root",
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}

	err = container.Start(process)
	if err != nil {
		return err
	}

	_, err = process.Wait()
	if err != nil {
		return err
	}

	return nil
}

func runContainerTimed(name string,
	args []string,
	wd string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	timeout time.Duration) error {
	var err error
	var id string

	id = path.Base(wd)

	// mount base rootfs with working directory
	rootfs := master_config.Rootfs
	lowerdir := rootfs
	upperdir, err := filepath.Abs(wd)
	if err != nil {
		return err
	}
	workdir, err := filepath.Abs(fmt.Sprintf("%s-%s", wd, "work"))
	if err != nil {
		return err
	}

	err = os.Mkdir(workdir, 0775)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer os.RemoveAll(workdir)
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		lowerdir, upperdir, workdir)
	err = syscall.Mount("overlay", upperdir, "overlay", syscall.MS_MGC_VAL,
		opts)
	if err != nil {
		return err
	}
	defer func() {
		err := syscall.Unmount(upperdir, 0)
		if err != nil {
			return
		}
	}()

	// set cgroup path
	var config configs.Config
	config = *master_config
	config.Cgroups.Path = fmt.Sprintf("%s/%s",
		config.Cgroups.Path, id)
	config.Rootfs = upperdir
	container, err := factory.Create(id, &config)
	if err != nil {
		return err
	}
	defer container.Destroy()

	sec := fmt.Sprintf("%d", int(timeout.Seconds()))
	args = append([]string{"timeout", "-k", "1", sec, name}, args...)
	process := &libcontainer.Process{
		Args:   args,
		Env:    []string{"PATH=/bin:/sbin:/usr/bin:/usr/sbin"},
		User:   "root",
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}

	start := time.Now()
	err = container.Start(process)
	if err != nil {
		return err
	}

	_, err = process.Wait()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			pst := ee.ProcessState
			if st, ok := pst.Sys().(syscall.WaitStatus); ok {
				if st.ExitStatus() == 124 {
					err = fmt.Errorf("program killed after %s",
						time.Now().Sub(start).String())
				}
			}
		}
		return err
	}

	return nil
}

func createWorkspace(c Compiler) (string, error) {
	var dir string
	var err error

	dir, err = ioutil.TempDir(DataStore, c.Name())
	if err != nil {
		return "", err
	}
	err = os.Chmod(dir, 0775)
	if err != nil {
		return "", err
	}

	return dir, nil
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

func setupWorkspace(c Compiler, sourcefile, code string) (string, error) {
	dir, err := createWorkspace(c)
	if err != nil {
		return "", err
	}
	srcpath := fmt.Sprintf("%s/%s", dir, sourcefile)
	err = writeSource(srcpath, code)
	if err != nil {
		return "", err
	}
	return dir, nil
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
