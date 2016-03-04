// Copyright 2016 Alex Fluter

package lotsawa

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fluter01/lotsawa/lang"
)

// Struct holds the compile request
type Request struct {
	// Time when the compiling started
	received time.Time
	// rpc args
	args *CompileArgs
	// channel to receive compiler's result
	chRes chan *lang.Result
}

// Compiler server
type CompilerServer struct {
	chReq  chan *Request
	chExit chan bool

	compilers map[string]lang.Compiler
}

func NewCompilerServer() *CompilerServer {
	s := new(CompilerServer)

	s.chReq = make(chan *Request)
	s.chExit = make(chan bool)
	s.compilers = make(map[string]lang.Compiler)

	// C defaults to C11
	s.AddCompiler("C", new(lang.C11))
	s.AddCompiler("C11", new(lang.C11))
	s.AddCompiler("C99", new(lang.C99))
	s.AddCompiler("C89", new(lang.C89))
	// alias sh to bash
	s.AddCompiler("sh", new(lang.Bash))
	s.AddCompiler("Bash", new(lang.Bash))

	return s
}

// manage compilers
func (s *CompilerServer) AddCompiler(name string, c lang.Compiler) {
	name = strings.ToUpper(name)
	s.compilers[name] = c
}

func (s *CompilerServer) DelCompiler(name string) {
	delete(s.compilers, name)
}

func (s *CompilerServer) GetCompiler(name string) lang.Compiler {
	name = strings.ToUpper(name)
	return s.compilers[name]
}

func (s *CompilerServer) ListCompiler() []string {
	var keys []string

	for key := range s.compilers {
		keys = append(keys, key)
	}
	return keys
}

func (s *CompilerServer) Init() error {
	var err error
	var cnt int

	for name, comp := range s.compilers {
		err = comp.Init()
		if err != nil {
			log.Printf("%s init failed: %s", name, err)
		} else {
			cnt++
		}
	}
	if cnt == 0 {
		return errors.New("Error: no compiler available to run")
	}

	fi, err := os.Stat(lang.DataStore)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("data store %s does not exists, creating now", lang.DataStore)
			err = os.MkdirAll(lang.DataStore, 0775)
			if err != nil {
				log.Printf("could not create %s: %s", lang.DataStore, err)
				return err
			} else {
				log.Printf("ok")
			}
		} else {
			log.Printf("could not access data store %s: %s", lang.DataStore, err)
			return err
		}
	} else {
		if !fi.IsDir() {
			log.Printf("data store(%s) exists, but is not directory", lang.DataStore)
			return errors.New("data store(" + lang.DataStore + ") exists, but is not directory")
		}
	}

	err = lang.InitContainer()
	if err != nil {
		log.Printf("container init failed: %s", err)
		log.Printf("will use unrestricted host execution")
	}
	return nil
}

func (s *CompilerServer) Loop() {
	var req *Request
	var stop bool = false
	for !stop {
		select {
		case req = <-s.chReq:
			s.handle(req)
			break
		case stop = <-s.chExit:
			break
		}
	}
}

func (s *CompilerServer) handle(req *Request) {
	var c lang.Compiler
	var res *lang.Result

	c = s.GetCompiler(req.args.Lang)

	if c == nil {
		res = &lang.Result{
			Error: "Language not supported.",
		}
	} else {
		res = c.Compile(req.args.Code)
	}

	req.chRes <- res
	return
}

func (s *CompilerServer) Submit(req *Request) {
	s.chReq <- req
}

func (s *CompilerServer) Run() {
	go s.Loop()
}
