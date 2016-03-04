// Copyright 2016 Alex Fluter

package lang

// Compile C code as C89
type C89 struct {
	CBase
}

func (c *C89) Name() string {
	return "GCC-C89"
}

func (c *C89) Init() error {
	if err := c.CBase.Init(); err != nil {
		return err
	}

	c.options = []string{
		"-Wextra",
		"-Wall",
		"-Wno-unused",
		"-pedantic",
		"-Wfloat-equal",
		"-Wshadow",
		"-std=c89",
		"-lm",
		"-Wfatal-errors",
		"-fsanitize=alignment,undefined"}
	c.prelude = `
#define _XOPEN_SOURCE 9001
#define __USE_XOPEN
#include <assert.h>
#include <ctype.h>
#include <errno.h>
#include <float.h>
#include <limits.h>
#include <locale.h>
#include <math.h>
#include <setjmp.h>
#include <signal.h>
#include <stdarg.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#line 1
`
	return nil
}

func (c *C89) Compile(code string) *Result {
	return c.compile(c, code, c.prelude)
}
