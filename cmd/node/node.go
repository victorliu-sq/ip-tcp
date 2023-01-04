package main

import (
	"log"
	"os"
	"tcpip/pkg/kernel"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Usage: node + inx file")
	}
	node := &kernel.Node{}
	node.Make(os.Args)
	// node.HandleCLI()
	node.ReceiveOpFromChan()
}
