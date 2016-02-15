// Copyright 2016 Alex Fluter

package lotsawa

import (
	"errors"
	"log"
	"strings"
	"time"
)

// Compiler is the interface that represents a compiler.
// Each compiler implements this interface, and registers
// itself to the compiler server to serve client's request.
type Compiler interface {
	// Returns the name of this compiler
	Name() string

	// Returns the version of this compiler
	Version() string

	// Do initialization work, for example, check whether files
	// required are present or not.
	// Returns an error if the compiler cannot function.
	// The compiler server will ignore this compiler if Init() failed.
	Init() error

	// Compile the code given as a string
	Compile(string) *Result
}

type Request struct {
	// Time when the compiling started
	received time.Time

	args  *CompileArgs
	chRes chan *Result
}

// Struct hold the compiling result
type Result struct {
	// Time when the compiling finished
	done time.Time

	// Whether the code has main function and can be executed
	main bool

	// The command line used to compile this piece of code
	cmd string

	// Compiler's standard error
	c_err string

	// Compiler's standard output
	c_out string

	// Errors during the compiling
	err string

	// The program's standard error, if compiled successfully and had main
	p_out string

	// The program's standard output, if compiled successfully and had main
	p_err string
}
