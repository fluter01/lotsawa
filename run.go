package lotsawa

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/specs"
)

func run(name string,
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
		log.Printf("Error run %s: %s", cmd, err)
		return err
	}
	return nil
}

func runTimed(name string,
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
		err := cmd.Run()
		if err != nil {
			log.Printf("Error run %s: %s", cmd, err)
		}
		chDone <- true
	}()

	var done bool = false
	select {
	case done = <-chDone:
		break
	case t := <-time.After(timeout):
		if !done {
			log.Println("program run timeout, killing")
			err = cmd.Process.Kill()
			if err != nil {
				log.Println("kill error:", err)
			}
			var st *os.ProcessState
			st, err = cmd.Process.Wait()
			if err != nil {
				log.Println("wait error:", err)
			}
			log.Println("state:", st)
			err = fmt.Errorf("program killed after %s",
				t.Sub(start).String())
		}
		break
	}
	close(chDone)

	return err
}

const (
	fspec = "config.json"
	fconf = "container.json"
)

var (
	spec          *specs.LinuxSpec
	master_config *configs.Config
	factory       libcontainer.Factory
	use_container bool
)

func initContainer() error {
	var err error
	spec, err = loadSpec(fspec)
	if err != nil {
		return err
	}

	factory, err = createFactory()
	if err != nil {
		return err
	}

	master_config, err = loadConfig(spec, fconf)
	if err != nil {
		return err
	}
	use_container = true
	return nil
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
	rootfs := spec.Root.Path
	lowerdir := rootfs
	upperdir, err := filepath.Abs(wd)
	if err != nil {
		log.Println("abs err:", wd, err)
		return err
	}
	workdir, err := filepath.Abs(fmt.Sprintf("%s-%s", wd, "work"))
	if err != nil {
		log.Println("abs err:", wd, err)
		return err
	}

	err = os.Mkdir(workdir, 0775)
	if err != nil && !os.IsExist(err) {
		log.Printf("failed to create workdir %s: %s", workdir, err)
		return err
	}
	defer os.RemoveAll(workdir)
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		lowerdir, upperdir, workdir)
	err = syscall.Mount("overlay", upperdir, "overlay", syscall.MS_MGC_VAL,
		opts)
	if err != nil {
		log.Println("mount failed:", err)
		return err
	}
	defer func() {
		err := syscall.Unmount(upperdir, 0)
		if err != nil {
			log.Printf("unmount error: %s\n", err)
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
		log.Printf("create %s error: %s\n", id, err)
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
		log.Printf("start container %s error: %s\n", id, err)
		return err
	}

	_, err = process.Wait()
	if err != nil {
		log.Printf("wait %s error: %s\n", id, err)
		return err
	}

	return nil
}
