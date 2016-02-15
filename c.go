// Copyright 2016 Alex Fluter

package lotsawa

import (
	"os"
)

type CCompiler struct {
}

func (c *CCompiler) Name() string {
	return "GCC"
}

func (c *CCompiler) Init() error {
	return nil
}

func (c *CCompiler) Compile(code string) *Result {
	return &Result{}
}
