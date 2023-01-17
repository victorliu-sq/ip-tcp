package kernel

import (
	"fmt"
	"log"
	"tcpip/pkg/proto"
	"time"
)

func (node *Node) HandleCmdRcvFile(nodeCLI *proto.NodeCLI) {
	// 1. Create a Listener
	port := nodeCLI.Val16
	listener, err := node.VListen(port)
	if err != nil {
		log.Fatalln(err)
	}
	// 2. Wait for a connection
	conn := listener.VAccept()
	// Wait until Establish
	for conn.GetState() != proto.ESTABLISH {
		time.Sleep(100 * time.Millisecond)
	}
	// 3. File Reader
	conn.FileReader(nodeCLI.Fd)
	// 4. Wait until receive one FIN
	for !conn.IsFINRcvd() {
		time.Sleep(100 * time.Millisecond)
	}
	// 5. Close the conn
	DPrintf("-------------------------------------------------------------")
	DPrintf("FIN has been acked: Send one FIN")
	DPrintf("-------------------------------------------------------------")
	conn.Close()
}

func (node *Node) HandleCmdSendFile(nodeCLI *proto.NodeCLI) {
	// 1. Dial to Listener
	remoteAddr, remotePort := nodeCLI.DestIP, nodeCLI.DestPort
	localAddr, ok := node.RT.FindSrcIPAddr(remoteAddr)
	if !ok {
		fmt.Errorf("Dest Addr does not exist\n")
		return
	}
	conn := node.ST.CreateConnSYNSENT(remoteAddr, localAddr, remotePort, node.NodeSegSendChan)
	// 2. Send SYN Segment
	conn.SendSeg3WHS_SYN()
	// Wait until Establish
	for conn.GetState() != proto.ESTABLISH {
		time.Sleep(100 * time.Millisecond)
	}
	// 3. File Writer
	conn.FileWriter(nodeCLI.Fd)

	// 4. Wait until all bytes have been acked
	for !conn.IsAllAcked() {
		time.Sleep(100 * time.Millisecond)
	}
	// 5. Close the conn
	DPrintf("-------------------------------------------------------------")
	DPrintf("All Bytes have been acked: Send one FIN")
	DPrintf("-------------------------------------------------------------")
	conn.Close()
}
