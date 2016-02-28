package lotsawa

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/specs"
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

func validateSpec(spec *specs.LinuxSpec) error {
	if spec.Process.Cwd == "" {
		return fmt.Errorf("Cwd property must not be empty")
	}
	if !filepath.IsAbs(spec.Process.Cwd) {
		return fmt.Errorf("Cwd must be an absolute path")
	}
	return nil
}

func loadSpec(path string) (spec *specs.LinuxSpec, err error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file %s not found", path)
		}
		return spec, err
	}
	defer f.Close()

	if err = json.NewDecoder(f).Decode(&spec); err != nil {
		return spec, err
	}
	return spec, validateSpec(spec)
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
