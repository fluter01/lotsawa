// Copyright 2016 Alex Fluter

package main

import (
	"fmt"

	"github.com/fluter01/lotsawa"
)

const addr = "127.0.0.1:1234"

func main() {
	// Synchronous call
	var err error

	var s *lotsawa.CompileServiceStub
	s, err = lotsawa.NewCompileServiceStub("tcp", addr)
	if err != nil {
		fmt.Println("Failed to dial rpc server:", err)
		return
	}
	var arg lotsawa.CompileArgs = lotsawa.CompileArgs{"abc", "c"}
	var res lotsawa.CompileReply
	err = s.Compile(&arg, &res)

	if err != nil {
		fmt.Println("Failed to call rpc service:", err)
		return
	}
	fmt.Println(res)
}
