// Copyright 2016 Alex Fluter

package lotsawa

import (
	"net/rpc"
	"time"
)

type CompileArgs struct {
	// The code to compile
	Code string

	// The language of the code
	Lang string
}

type CompileReply struct {
	// The command line used to compile this piece of code
	Cmd string
	// Whether the code has main function and can be executed
	Error string
	// Time took to compile and run the program
	Time time.Duration
	// Compiler's standard output
	C_Output string
	// Compiler's standard error
	C_Error string
	// The program's standard error, if compiled successfully and run
	P_Output string
	// The program's standard output, if compiled successfully and run
	P_Error string
}

// RPC service
type CompileService struct {
	server *CompilerServer
}

func NewCompileService(server *CompilerServer) *CompileService {
	return &CompileService{
		server: server,
	}
}

func (c *CompileService) Compile(args *CompileArgs, reply *CompileReply) error {
	req := &Request{
		received: time.Now(),
		args:     args,
		chRes:    make(chan *Result),
	}

	c.server.Submit(req)

	res := <-req.chRes

	*reply = res.CompileReply
	reply.Time = time.Now().Sub(req.received)

	close(req.chRes)
	return nil
}

// RPC stub for the client
type CompileServiceStub struct {
	client *rpc.Client
}

func NewCompileServiceStub(net, addr string) (*CompileServiceStub, error) {
	var err error
	s := new(CompileServiceStub)
	s.client, err = rpc.Dial(net, addr)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (c *CompileServiceStub) Compile(args *CompileArgs, reply *CompileReply) error {
	var err error

	err = c.client.Call("CompileService.Compile", args, reply)
	return err
}

func (c *CompileServiceStub) Close() error {
	return c.client.Close()
}
