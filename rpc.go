// Copyright 2016 Alex Fluter

package lotsawa

import (
	"net"
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
	Main bool
	// Errors during the compiling
	Error string
	// Time took to compile and run the program
	Time time.Duration
	// Compiler's standard output
	C_Output string
	// Compiler's standard error
	C_Error string
	// The program's standard error, if compiled successfully and had main
	P_Output string
	// The program's standard output, if compiled successfully and had main
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
	res.done = time.Now()

	reply.Time = res.done.Sub(req.received)

	reply.Error, reply.Cmd, reply.Main = res.err, res.cmd, res.main
	reply.C_Output, reply.C_Error = res.c_out, res.c_err
	if res.main {
		reply.P_Output, reply.P_Error = res.p_out, res.p_err
	}

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

// RPC server
type RpcServer struct {
	svr    *rpc.Server
	l      *net.TCPListener
	addr   *net.TCPAddr
	server *Server
}

func NewRpcServer(server *Server, addr string) (*RpcServer, error) {
	var err error
	s := new(RpcServer)
	s.svr = rpc.NewServer()
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	s.addr = tcpAddr
	s.server = server

	return s, nil
}

func (s *RpcServer) Init() error {
	l, err := net.ListenTCP("tcp", s.addr)
	if err != nil {
		return err
	}
	s.l = l
	s.svr.Register(NewCompileService(s.server.compSvr))
	return nil
}

func (s *RpcServer) Run() {
	go s.svr.Accept(s.l)
}

func (s *RpcServer) Wait() {
	s.svr.Accept(s.l)
}
