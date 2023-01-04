package kernel

import (
	"tcpip/pkg/myDebug"
	"tcpip/pkg/network"
	"tcpip/pkg/proto"
	"tcpip/pkg/tcp"
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
	socketTable *tcp.SocketTable
	segRecvChan chan *proto.Segment //seg received from the network/router(PROTO:6)
	segSendChan chan *proto.Segment //seg to be sent from normal socket
	blockCLI    bool
}

func (node *Node) Make(args []string) {
	myDebug.InitDebugger()
	// Initialize Channel
	node.NodeCLIChan = make(chan *proto.NodeCLI)
	node.NodeBCChan = make(chan *proto.NodeBC)
	node.NodeExChan = make(chan *proto.NodeEx)
	node.NodePktOpChan = make(chan *proto.NodePktOp)

	node.socketTable = tcp.NewSocketTable()
	node.segRecvChan = make(chan *proto.Segment)
	node.segSendChan = make(chan *proto.Segment, 100)

	node.RT = &network.RoutingTable{}
	node.RT.Make(args, node.NodePktOpChan, node.NodeExChan, node.segRecvChan)

	// Receive CLI
	go node.ScanClI()
	// Broadcast RIP Request once
	go node.RIPReqDaemon()
	// Broadcast RIP Resp periodically
	go node.RIPRespDaemon()

	go node.handleTCPSegment()
}
