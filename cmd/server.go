// Copyright 2016 Alex Fluter

package main

import (
	"fmt"

	"github.com/fluter01/lotsawa"
)

const addr = "127.0.0.1:1234"

func main() {
	var err error
	var s *lotsawa.Server

	s, err = lotsawa.NewServer(addr)
	if err != nil {
		fmt.Println("Failed to create server")
		return
	}

	fmt.Println("Server running on", addr)
	s.Wait()
}
