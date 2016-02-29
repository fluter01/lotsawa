// Copyright 2016 Alex Fluter

package lotsawa

import (
	"net"
	"net/rpc"
)

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
