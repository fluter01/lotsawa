// Copyright 2016 Alex Fluter

package lotsawa

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"
)

const DataStore = "store"

// Struct holds the compile request
type Request struct {
	// Time when the compiling started
	received time.Time
	// rpc args
	args *CompileArgs
	// channel to receive compiler's result
	chRes chan *Result
}

// Compiler server
type CompilerServer struct {
	chReq  chan *Request
	chExit chan bool

	compilers map[string]Compiler
}

func NewCompilerServer() *CompilerServer {
	s := new(CompilerServer)

	s.chReq = make(chan *Request)
	s.chExit = make(chan bool)
	s.compilers = make(map[string]Compiler)

	s.AddCompiler("C", new(CCompiler))

	return s
}

// manage compilers
func (s *CompilerServer) AddCompiler(name string, c Compiler) {
	name = strings.ToUpper(name)
	s.compilers[name] = c
}

func (s *CompilerServer) DelCompiler(name string) {
	delete(s.compilers, name)
}

func (s *CompilerServer) GetCompiler(name string) Compiler {
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

	fi, err := os.Stat(DataStore)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("data store %s does not exists, creating now", DataStore)
			err = os.MkdirAll(DataStore, 0775)
			if err != nil {
				log.Printf("could not create %s: %s", DataStore, err)
				return err
			} else {
				log.Printf("ok")
			}
		} else {
			log.Printf("could not access data store %s: %s", DataStore, err)
			return err
		}
	} else {
		if !fi.IsDir() {
			log.Printf("data store(%s) exists, but is not directory", DataStore)
			return errors.New("data store(" + DataStore + ") exists, but is not directory")
		}
	}

	err = initContainer()
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
	var lang string
	var c Compiler
	var res *Result

	lang = req.args.Lang
	c = s.GetCompiler(lang)

	if c == nil {
		cr := CompileReply{
			Error: "Language not supported.",
		}
		res = &Result{
			CompileReply: cr,
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
