// Copyright 2016 Alex Fluter

package lang

// Compile as C11
type C11 struct {
	CBase
}

func (c *C11) Name() string {
	return "GCC-C11"
}

func (c *C11) Init() error {
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
		"-std=c11",
		"-lm",
		"-Wfatal-errors",
		"-fsanitize=alignment,undefined"}
	c.prelude = `
#define _XOPEN_SOURCE 9001
#define __USE_XOPEN
// list of all C11 headers
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
#include <stdalign.h>
#include <stdarg.h>
#include <stdatomic.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdnoreturn.h>
#include <string.h>
#include <tgmath.h>
#if __STDC_NO_THREADS__ != 1
#include <threads.h>
#endif
#include <time.h>
#include <uchar.h>
#include <wchar.h>
#include <wctype.h>
// unix headers
#include <unistd.h>
#include <sys/types.h>

#line 1
`
	return nil
}

func (c *C11) Compile(code string) *Result {
	return c.compile(code, c.prelude)
}
