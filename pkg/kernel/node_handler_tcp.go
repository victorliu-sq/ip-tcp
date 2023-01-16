package kernel

import (
	"fmt"
	"tcpip/pkg/proto"
	"tcpip/pkg/transport"

	"github.com/google/netstack/tcpip/header"
)

// ***********************************************************************************
// *** FUNCTIONS FOR LISTENING SOCKETS (VTCPListener) ***
func (node *Node) VListen(port uint16) (*transport.VTCPListener, error) {
	if _, ok := node.ST.Port2Listener(port); ok {
		return nil, fmt.Errorf("Local port exists in Socket Table\n")
	}
	listener := node.ST.CreateListener(port, node.NodeSegSendChan)
	return listener, nil
}

// ***********************************************************************************
// *** FUNCTIONS FOR NORMAL SOCKETS (VTCPConn) ***
func (node *Node) VConnect(remoteAddr string, remotePort uint16) (*transport.VTCPConn, error) {
	localAddr, found := node.RT.FindSrcIPAddr(remoteAddr)
	if !found {
		return nil, fmt.Errorf("Dest Addr does not exist\n")
	}
	conn := node.ST.CreateConnSYNSENT(remoteAddr, localAddr, remotePort, node.NodeSegSendChan)
	// Send SYN Segment
	conn.SendSeg3WHS_SYN()
	return conn, nil
}

func (node *Node) HandleSegmentToSend(segment *proto.Segment) {
	hdr := segment.TCPhdr
	payload := segment.Payload
	tcpHeaderBytes := make(header.TCP, proto.TcpHeaderLen)
	tcpHeaderBytes.Encode(hdr)
	iPayload := make([]byte, 0, len(tcpHeaderBytes)+len(payload))
	iPayload = append(iPayload, tcpHeaderBytes...)
	iPayload = append(iPayload, []byte(payload)...)
	// proto.PrintHex(iPayload)
	node.RT.SendTCPPacket(segment.IPhdr.Src.String(), segment.IPhdr.Dst.String(), string(iPayload))
}

func (node *Node) HandleSegmentReceived(segment *proto.Segment) {
	// 1. Try to Send to Normal Conn
	tuple := segment.FormTuple()
	conn, ok := node.ST.Tuple2Conn(tuple)
	if ok {
		// fmt.Println("Sent to conn")
		conn.ConnSegRcvChan <- segment
		return
	}

	// 2. Try to Send to Listener of DstPort if No Corresponding Normal Conn Exists
	port := segment.TCPhdr.DstPort
	listener, ok := node.ST.Port2Listener(port)
	if ok {
		// fmt.Println("Sent to Listener")
		listener.ListenerSegRcvChan <- segment
		return
	}
}

// *****************************************************************************************
// Handle Cmd Send Segment
func (node *Node) HandleCmdSendSegment(nodeCLI *proto.NodeCLI) {
	socketID := nodeCLI.Val16
	conn, ok := node.ST.ID2Conn(socketID)
	if !ok {
		fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
		return
	}
	go conn.WriteBytesToSNDLoop(nodeCLI.Bytes)
}

// *****************************************************************************************
// Handle Cmd Rcv Segment
func (node *Node) HandleCmdRcvSegment(nodeCLI *proto.NodeCLI) {
	socketID := nodeCLI.Val16
	conn, ok := node.ST.ID2Conn(socketID)
	if !ok {
		fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
		return
	}
	numBytes := nodeCLI.Val32
	conn.ReadBytesFromRCVLoop(numBytes)
}

// *****************************************************************************************
// Handle Cmd Close
func (node *Node) HandleCmdClose(nodeCLI *proto.NodeCLI) {
	socketID := nodeCLI.Val16
	conn, ok := node.ST.ID2Conn(socketID)
	if !ok {
		fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
		return
	}
	// Close the connection
	conn.Close()
}
