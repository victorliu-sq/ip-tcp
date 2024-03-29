package kernel

import (
	"fmt"
	"log"
	"os"
	"tcpip/pkg/proto"
)

// Output the data of CLI
func (node *Node) ReceiveOpFromChan() {
	fmt.Printf("> ")
	for {
		select {
		case nodeCLI := <-node.NodeCLIChan:
			// fmt.Println(nodeCLI)
			node.HandleNodeCLI(nodeCLI)
		case nodeBC := <-node.NodeBCChan:
			// fmt.Println(nodeBC)
			node.HandleNodeBC(nodeBC)
		case nodeEx := <-node.NodeExChan:
			// fmt.Println(nodeEx)
			node.HandleNodeEx(nodeEx)
		case nodePktOp := <-node.NodePktOpChan:
			// fmt.Println(nodePktOp)
			node.HandleNodePktOp(nodePktOp)
		case segment := <-node.NodeSegSendChan:
			// fmt.Println("Send one Segment")
			node.HandleSegmentToSend(segment)
		case segment := <-node.NodeSegRecvChan:
			// fmt.Println("Receives one Segment")
			node.HandleSegmentReceived(segment)
		}
	}
}

func (node *Node) HandleNodeCLI(nodeCLI *proto.NodeCLI) {
	switch nodeCLI.CLIType {
	// CLI
	case proto.CLI_LI:
		node.HandlePrintInterfaces()
		fmt.Printf("> ")
	case proto.CLI_SETUP:
		node.HandleSetUp(nodeCLI.ID)
		fmt.Printf("> ")
	case proto.CLI_SETDOWN:
		node.HandleSetDown(nodeCLI.ID)
		fmt.Printf("> ")
	case proto.CLI_QUIT:
		node.HandleQuit()
		fmt.Printf("> ")
	case proto.CLI_LR:
		node.HandlePrintRoutes()
		fmt.Printf("> ")
	case proto.MESSAGE_SENDPKT:
		node.HandleSendPacket(nodeCLI.DestIP, nodeCLI.ProtoID, nodeCLI.Msg)
		fmt.Printf("> ")
	case proto.CLI_LIFILE:
		node.HandlePrintInterfacesToFile(nodeCLI.Filename)
		fmt.Printf("> ")
	case proto.CLI_LRFILE:
		node.HandlePrintRoutesToFile(nodeCLI.Filename)
		fmt.Printf("> ")
	case proto.CLI_CREATELISTENER:
		// node.handleCreateListener(nodeCLI)
		node.HandleCreateListener(nodeCLI)
		fmt.Printf("> ")
	case proto.CLI_LS:
		node.HandlePrintSockets()
		fmt.Printf("> ")
	case proto.CLI_CREATECONN:
		node.HandleCreateConn(nodeCLI)
		fmt.Printf("> ")
	case proto.CLI_SENDSEGMENT:
		node.HandleCmdSendSegment(nodeCLI)
		fmt.Printf("> ")
	case proto.CLI_RECVSEGMENT:
		go node.HandleCmdRcvSegment(nodeCLI)
		fmt.Printf("> ")
	case proto.CLI_CLOSE:
		node.HandleCmdClose(nodeCLI)
		fmt.Printf("> ")
	case proto.CLI_RCVFILE:
		go node.HandleCmdRcvFile(nodeCLI)
		fmt.Printf("> ")
	case proto.CLI_SNDFILE:
		go node.HandleCmdSendFile(nodeCLI)
		fmt.Printf("> ")
	}
}

func (node *Node) HandleNodeBC(nodeBC *proto.NodeBC) {
	switch nodeBC.OpType {
	case proto.MESSAGE_BCRIPREQ:
		node.HandleBroadcastRIPReq()
	case proto.MESSAGE_BCRIPRESP:
		node.HandleBroadcastRIPResp()
	}
}

func (node *Node) HandleNodeEx(nodeEx *proto.NodeEx) {
	switch nodeEx.OpType {
	case proto.MESSAGE_ROUTEEX:
		node.HandleRouteEx(nodeEx.DestIP)
	}
}

func (node *Node) HandleNodePktOp(nodePktOp *proto.NodePktOp) {
	switch nodePktOp.OpType {
	case proto.MESSAGE_REVPKT:
		node.HandleReceivePacket(nodePktOp.Bytes.([]byte), nodePktOp.DestIP)
	}
}

// ***********************************************************************************
// Handle CLI
func (node *Node) HandlePrintInterfaces() {
	node.RT.PrintInterfaces()
}

func (node *Node) HandlePrintInterfacesToFile(filename string) {
	node.RT.PrintInterfacesToFile(filename)
}

func (node *Node) HandleSetUp(id uint8) {
	node.RT.SetUp(id)
}

func (node *Node) HandleSetDown(id uint8) {
	node.RT.SetDown(id)
}

func (node *Node) HandleQuit() {
	os.Exit(0)
}

func (node *Node) HandlePrintRoutes() {
	node.RT.PrintRoutes()
}

func (node *Node) HandlePrintRoutesToFile(filename string) {
	node.RT.PrintRoutesToFile(filename)
}

func (node *Node) HandleSendPacket(destIP string, protoID int, msg string) {
	node.RT.SendPacket(destIP, msg)
}

func (node *Node) HandlePrintSockets() {
	node.ST.PrintSockets()
}

func (node *Node) HandleCreateListener(nodeCLI *proto.NodeCLI) {
	port := nodeCLI.Val16
	listener, err := node.VListen(port)
	if err != nil {
		log.Fatalln(err)
	}
	go listener.VAcceptLoop()
}

func (node *Node) HandleCreateConn(nodeCLI *proto.NodeCLI) {
	_, err := node.VConnect(nodeCLI.DestIP, nodeCLI.DestPort)
	if err != nil {
		log.Fatalln(err)
	}
}

// ***********************************************************************************
// Handle BroadcastRIP
func (node *Node) HandleBroadcastRIPReq() {
	// fmt.Println("Try to broadcast RIP Req")
	node.RT.BroadcastRIPReq()
}

func (node *Node) HandleBroadcastRIPResp() {
	// fmt.Println("Try to broadcast RIP Resp")
	node.RT.BroadcastRIPResp()
}

// ***********************************************************************************
// Handle Expired Route
func (node *Node) HandleRouteEx(destIP string) {
	node.RT.CheckRouteEx(destIP)
}
