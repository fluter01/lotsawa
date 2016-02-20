package lotsawa

import (
	"fmt"
	"os"
	"testing"
)

var s *Server = nil

const addr = "127.0.0.1:1234"

func TestFoo(t *testing.T) {
	fmt.Println(os.TempDir())
}

func startServer() {
	if s != nil {
		return
	}
	var err error

	s, err = NewServer(addr)
	if err != nil {
		fmt.Println("Failed to create server:", err)
		return
	}

	fmt.Println("Server running on", addr)
	go s.Wait()
}

func getClient(t *testing.T) *CompileServiceStub {
	var s *CompileServiceStub
	var err error
	s, err = NewCompileServiceStub("tcp", addr)
	if err != nil {
		t.Fatal("Failed to dial rpc server:", err)
		return nil
	}
	return s
}

func testRun(code, lang string, t *testing.T) *CompileReply {
	var err error
	startServer()

	s := getClient(t)
	defer s.Close()

	var arg CompileArgs = CompileArgs{code, lang}
	var res CompileReply
	err = s.Compile(&arg, &res)

	if err != nil {
		t.Error(err)
		return nil
	}
	return &res
}

func (r *CompileReply) String() string {
	return fmt.Sprintf("Cmd: %s\nExec:%t, Took:%s\nError:%s\nCompile:%s|%s\nRun:%s|%s\n",
		r.Cmd,
		r.Main,
		r.Time,
		r.Error,
		r.C_Output,
		r.C_Error,
		r.P_Output,
		r.P_Error)
}

func TestCompile1(t *testing.T) {
	res := testRun("abc", "c", t)

	t.Log(res)
}

func TestCompile2(t *testing.T) {
	var code string = `
	#include <stdio.h>
	int main(void) {
		puts("hello");
		return 0;
	}
	`
	res := testRun(code, "c", t)

	t.Log(res)
}

func TestCompile3(t *testing.T) {
	var code string = `
	#include <stdio.h>
	int foo(void) {
		puts("foo");
	}
	`
	res := testRun(code, "c", t)

	t.Log(res)
}

func TestCompile4(t *testing.T) {
	var code string = `
	#include <stdio.h>
	int foo(void) {
		puts("foo");
	}
	int main(void) {
		foo();
		return 0;
	}
	`
	res := testRun(code, "c", t)

	t.Log(res)
}

func TestCompile5(t *testing.T) {
	var code string = `
	#include <stdio.h>
	int main(void) {
		fprintf(stdout, "output to stdout\n");
		fprintf(stderr, "output to stderr\n");
		return 0;
	}
	`
	res := testRun(code, "c", t)

	t.Log(res)
}

func TestCompile6(t *testing.T) {
	var code string = `
	#include <stdio.h>
	int main(int argc, char *argv[]) {
		int *p = 3;
		fprintf(stdout, "output to stdout\n");
		fprintf(stderr, "output to stderr\n");
		return 0;
	}
	`
	res := testRun(code, "c", t)

	t.Log(res)
}

func TestCompile7(t *testing.T) {
	var code string = `
	#include <stdio.h>
	int main(int argc, char *argv[]) {
		int *p = 3;
		*p = 1;
		fprintf(stdout, "output to stdout\n");
		fprintf(stderr, "output to stderr\n");
		return 0;
	}
	`
	res := testRun(code, "c", t)

	t.Log(res)
}

func TestCompile8(t *testing.T) {
	var code string = `
	#include <stdio.h>
	int main(int argc, char *argv[]) {
		while (1) {
			printf("+");
			;
		}
		return 0;
	}
	`
	res := testRun(code, "c", t)

	t.Log(res)
}
