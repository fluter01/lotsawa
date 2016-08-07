// Copyright 2016 Alex Fluter

package lang

import (
	"encoding/json"
	"fmt"
	"io"
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
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	fconf     = "libcontainer.json"
	runc_root = "/run/lotsawa/runc"
)

var (
	master_config *configs.Config
	factory       libcontainer.Factory
	use_container bool
)

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

	err = os.MkdirAll(runc_root, 0700)
	if err != nil {
		return err
	}
	err = syscall.Access(runc_root, 0x7)
	if err != nil {
		return err
	}

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
	var spec *specs.Spec
	if err = json.NewDecoder(f).Decode(&spec); err != nil {
		return nil, err
	}

	var config *configs.Config
	config, err = specconv.CreateLibcontainerConfig(&specconv.CreateOpts{
		CgroupName:       "",
		UseSystemdCgroup: false,
		NoPivotRoot:      true,
		NoNewKeyring:     true,
		Spec:             spec,
	})
	return config, nil
}

func createFactory() (libcontainer.Factory, error) {
	abs, err := filepath.Abs(runc_root)
	if err != nil {
		return nil, err
	}
	return libcontainer.New(abs, libcontainer.Cgroupfs, func(l *libcontainer.LinuxFactory) error {
		l.CriuPath = "criu"
		return nil
	})
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

	err = container.Run(process)
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

	sec := fmt.Sprintf("%d", int(timeout.Seconds()))
	args = append([]string{"-k", "1", sec, name}, args...)
	err := runContainer("timeout",
		args,
		wd,
		stdin,
		stdout,
		stderr)

	start := time.Now()
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
