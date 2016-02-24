package lotsawa

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
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

	return err
}

func runInSandbox() {
}
