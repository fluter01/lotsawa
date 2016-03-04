// Copyright 2016 Alex Fluter

package lang

// Compile C code as C99
type C99 struct {
	CBase
}

func (c *C99) Name() string {
	return "GCC-C99"
}

func (c *C99) Init() error {
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
		"-std=c99",
		"-lm",
		"-Wfatal-errors",
		"-fsanitize=alignment,undefined"}
	c.prelude = `
#define _XOPEN_SOURCE 9001
#define __USE_XOPEN
#include <assert.h>
#include <complex.h>
#include <ctype.h>
#include <errno.h>
#include <fenv.h>
#include <float.h>
#include <inttypes.h>
#include <limits.h>
#include <locale.h>
#include <math.h>
#include <setjmp.h>
#include <signal.h>
#include <stdarg.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <tgmath.h>
#include <time.h>
#include <wchar.h>
#include <wctype.h>

#line 1
`
	return nil
}

func (c *C99) Compile(code string) *Result {
	return c.compile(c, code, c.prelude)
}
