// Copyright 2016 Alex Fluter

package lang

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

// Struct hold the compiling result
type Result struct {
	// The command line used to compile this piece of code
	Cmd string
	// Whether the code has main function and can be executed
	Error string
	// Compiler's standard output
	C_Output string
	// Compiler's standard error
	C_Error string
	// The program's standard error, if compiled successfully and run
	P_Output string
	// The program's standard output, if compiled successfully and run
	P_Error string
}

// The directory to store the source and compiled binary files.
// Any produced files by the program are also placed under it.
const DataStore = "store"
