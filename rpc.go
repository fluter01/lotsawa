// Copyright 2016 Alex Fluter

package lotsawa

import (
	"net"
	"net/rpc"
	"time"
)

type CompileArgs struct {
	Code string
	Lang string
}

type CompileReply struct {
	Cmd      string
	Main     bool
	Error    string
	Time     time.Duration
	C_Output string
	C_Error  string
	P_Output string
	P_Error  string
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
