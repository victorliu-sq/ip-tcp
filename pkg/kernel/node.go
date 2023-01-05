package kernel

import (
	"tcpip/pkg/network"
	"tcpip/pkg/proto"
	"tcpip/pkg/transport"
)

// The driver program
type Node struct {
	// Network Layer
	NodeCLIChan   chan *proto.NodeCLI   // Receive CLI from user
	NodeBCChan    chan *proto.NodeBC    // Broadcast RIP
	NodeExChan    chan *proto.NodeEx    // Handle expiration of route
	NodePktOpChan chan *proto.NodePktOp // Receive msg from link interface
	RT            *network.RoutingTable
	// Transport Layer
	ST              *transport.SocketTable
	NodeSegRecvChan chan *proto.Segment //seg received from the network/router(PROTO:6)
	NodeSegSendChan chan *proto.Segment //seg to be sent from normal socket
	// blockCLI        bool
}

func (node *Node) Make(args []string) {
	// Initialize Channel
	node.NodeCLIChan = make(chan *proto.NodeCLI)
	node.NodeBCChan = make(chan *proto.NodeBC)
	node.NodeExChan = make(chan *proto.NodeEx)
	node.NodePktOpChan = make(chan *proto.NodePktOp)
	node.NodeSegRecvChan = make(chan *proto.Segment)
	node.NodeSegSendChan = make(chan *proto.Segment)

	// Network
	node.RT = &network.RoutingTable{}
	// notice that Routing table needs channel of NodeS
	node.RT.Make(args, node.NodePktOpChan, node.NodeExChan, node.NodeSegRecvChan)

	// transport
	node.ST = transport.NewSocketTable()

	// Receive CLI
	go node.ScanClI()
	// Broadcast RIP Request once
	go node.RIPReqDaemon()
	// Broadcast RIP Resp periodically
	go node.RIPRespDaemon()
}
