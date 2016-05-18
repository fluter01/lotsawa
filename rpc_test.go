package lotsawa

import (
	"fmt"
	"sync"
	"testing"
)

const addr = "127.0.0.1:1234"

func (r *CompileReply) String() string {
	return fmt.Sprintf("Cmd: %s\nTook:%s\nError:%s\nCompile:%s|%s\nRun:%s|%s\n",
		r.Cmd,
		r.Time,
		r.Error,
		r.C_Output,
		r.C_Error,
		r.P_Output,
		r.P_Error)
}

var testData []CompileArgs = []CompileArgs{
	{`abc`, "c"},
	{`
	#include <stdio.h>
	int main(void) {
		puts("hello");
		printf("%d\n", __STDC_NO_THREADS__);
		return 0;
	}
	`, "c"},
	{`
	#include <stdio.h>
	int foo(void) {
		puts("foo");
	}
	`, "c"},
	{`
	#include <stdio.h>
	int foo(void) {
		puts("foo");
	}
	int main(void) {
		foo();
		return 0;
	}
	`, "c"},
	{`
	#include <stdio.h>
	int main(void) {
		fprintf(stdout, "output to stdout\n");
		fprintf(stderr, "output to stderr\n");
		return 0;
	}
	`, "c"},
	{`#include <stdio.h>
	int main(int argc, char *argv[]) {
		int *p = 3;
		fprintf(stdout, "output to stdout\n");
		fprintf(stderr, "output to stderr\n");
		return 0;
	}
	`, "c"},
	{`
	#include <stdio.h>
	int main(int argc, char *argv[]) {
		int *p = 3;
		*p = 1;
		fprintf(stdout, "output to stdout\n");
		fprintf(stderr, "output to stderr\n");
		return 0;
	}
	`, "c"},
	{`
	#include <stdio.h>
	int main(int argc, char *argv[]) {
		while (1) {
			printf("+");
			;
		}
		return 0;
	}
	`, "c"},
	{`
	#include <stdio.h>
	int main(int argc, char *argv[]) {
		FILE *fp = fopen("foo.txt", "w");
		perror("fopen");
		fprintf(fp, "hello\n");
		fclose(fp);
		return 0;
	}
	`, "c"},
	{`
	pwd
	uname -a
	`, "sh"},
	{`
		fmt.Println("hello lotsawa go");
	`, "go"},
	{`
		var i int
		var j int
		i = 3
		j = i + 5
		fmt.Println("i =", i, "j =", j)
	`, "go"},
	{`package main
import "fmt"
func main() {
	foo()
}
func foo() {
	fmt.Println("in foo")
}`, "go"},
	{`panic("foo")`, "go"},
	{`i = 3`, "go"},
	{`i := 3`, "go"},
	{`type S struct {
	a int
	}

	func main() {
	s := S{3}
	fmt.Println(s.a)
	}`, "go"},
	{`s := "hello"
	fmt.Println(s)`, "go"}, // frag
	{`func main() {
	fmt.Println("Hello")
}`, "go"}, // func
	{`package main
	func main() {
		fmt.Println("Hello")
	}`, "go"}, // package
	{`s := "hello"`, "go"}, // frag
}

func startServer(t *testing.T, exit chan bool) *Server {
	var err error

	s, err := NewServer(addr)
	if err != nil {
		t.Fatal("Failed to create server:", err)
		return nil
	}

	go func() {
		s.Wait()
		exit <- true
	}()
	return s
}

func stopServer(s *Server) {
	s.Stop()
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

func TestList(t *testing.T) {
	var err error
	var exit chan bool = make(chan bool)
	s := startServer(t, exit)
	defer func() {
		stopServer(s)
		<-exit
	}()

	c := getClient(t)
	defer c.Close()

	var res ListReply
	err = c.List(struct{}{}, &res)
	if err != nil {
		t.Error(err)
	}
	if len(res.Compilers) < 1 {
		t.Fail()
	}
}

func TestCompile(t *testing.T) {
	var err error
	var exit chan bool = make(chan bool)
	s := startServer(t, exit)
	defer func() {
		stopServer(s)
		<-exit
	}()

	c := getClient(t)
	defer c.Close()

	t.Log("Running", len(testData), "compile cases")
	for _, arg := range testData {
		var res CompileReply
		err = c.Compile(&arg, &res)

		if err != nil {
			t.Error(err)
		}
		t.Log(&res)
	}
}

func TestBench(t *testing.T) {
	t.SkipNow()
	var wg sync.WaitGroup

	for j := 0; j < 10; j++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 10; i++ {
				var code string = `
	#include <stdio.h>
	int main(int argc, char *argv[]) {
		puts("in prog %d");
		return 0;
	}
	`
				code = fmt.Sprintf(code, i)
				s := getClient(t)
				defer s.Close()
				var res CompileReply
				arg := CompileArgs{code, "c"}
				err := s.Compile(&arg, &res)

				if err != nil {
					t.Error(err)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
