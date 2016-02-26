package lotsawa

import (
	"bytes"
	"testing"
)

func TestContainer(t *testing.T) {
	err := initContainer()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = runContainer("hostname",
		[]string{"-a"},
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
