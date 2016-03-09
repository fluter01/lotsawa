// Copyright 2016 Alex Fluter

package lotsawa

import "log"

type Server struct {
	compSvr *CompilerServer
	rpcSvr  *RpcServer
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func NewServer(addr string) (*Server, error) {
	var err error
	s := new(Server)

	s.compSvr = NewCompilerServer()

	err = s.compSvr.Init()
	if err != nil {
		return nil, err
	}

	s.rpcSvr, err = NewRpcServer(s, addr)
	if err != nil {
		return nil, err
	}

	err = s.rpcSvr.Init()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Stop() {
	s.compSvr.Stop()
	s.rpcSvr.Stop()
}

func (s *Server) Wait() {
	s.compSvr.Run()
	s.rpcSvr.Wait()
}

func (s *Server) Run() {
	s.compSvr.Run()
	s.rpcSvr.Run()
}
