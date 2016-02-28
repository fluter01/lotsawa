package lotsawa

import (
	"bytes"
	"testing"
	"time"
)

func TestContainer(t *testing.T) {
	err := initContainer()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runContainer("ls",
		[]string{"-a", "-l"},
		"store/GCC775507155",
		nil,
		&stdout,
		&stderr)
	if err != nil {
		t.Error(err)
	}

	t.Log(stdout.String())
	t.Log(stderr.String())
}

func TestContainer2(t *testing.T) {
	err := initContainer()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runContainerTimed("sleep",
		[]string{"10"},
		"store/GCC775507155",
		nil,
		&stdout,
		&stderr,
		3*time.Second)
	if err != nil {
		t.Error(err)
	}

	t.Log(stdout.String())
	t.Log(stderr.String())
}
