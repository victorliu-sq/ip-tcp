package kernel

import (
	"fmt"
	"tcpip/pkg/proto"
	"tcpip/pkg/transport"

	"github.com/google/netstack/tcpip/header"
)

func (node *Node) handleTCPSegment() {
	// for {
	// 	segment := <-node.segRecvChan
	// 	tuple := segment.FormTuple()
	// 	if conn := node.socketTable.FindConn(tuple); conn != nil {
	// 		conn.SegRcvChan <- segment
	// 		continue
	// 	}
	// 	dstPort := segment.TCPhdr.DstPort
	// 	listener := node.socketTable.FindListener(dstPort)
	// 	if listener != nil {
	// 		listener.SegRcvChan <- segment
	// 	}
	// }
}

// *****************************************************************************************
// Handle Create Listener
// func (node *Node) handleCreateListener(msg *proto.NodeCLI) {
// 	val, _ := strconv.Atoi(msg.Msg)
// 	port := uint16(val)
// 	if node.socketTable.FindListener(port) != nil {
// 		fmt.Printf("Cannot assign requested address\n")
// 	} else {
// 		listener := node.socketTable.OfferListener(port)
// 		go node.NodeAcceptLoop(listener, false)
// 	}
// }

// // a port -> listener  -> go node.acceptConn(listener)
// func (node *Node) NodeAcceptLoop(listener *tcp.VTCPListener, oneTime bool) {
// 	for {
// 		conn, err := listener.VAccept()
// 		if err != nil {
// 			continue
// 		}
// 		conn.NodeSegSendChan = node.segSendChan
// 		conn.CLIChan = node.NodeCLIChan
// 		conn.RcvBuf = tcp.NewRecvBuffer(conn.GetAck(), tcp.DEFAULTWINDOWSIZE)
// 		node.socketTable.OfferConn(conn)
// 		// Recv SYN
// 		go conn.SynRev()
// 		if oneTime {
// 			listener.CancelChan <- true
// 			for listener.ConnInQueue > 0 {
// 				conn := <-listener.ConnQueue
// 				node.socketTable.DeleteSocket(conn.ID)
// 			}
// 			cli := <-listener.CLIChan
// 			node.socketTable.DeleteSocket(listener.ID)
// 			go conn.RetrivFile(cli.Fd)
// 		}
// 	}
// }

// // *****************************************************************************************
// // Handle Create Conn
// func (node *Node) HandleCreateConn(nodeCLI *proto.NodeCLI) *tcp.VTCPConn {
// 	// Create a Normal Socket
// 	srcIP := node.RT.FindSrcIPAddr(nodeCLI.DestIP)
// 	if srcIP == "no" {
// 		fmt.Println("v_connect() error: No route to host")
// 		return nil
// 	}
// 	conn := tcp.NewNormalSocket(0, nodeCLI.DestPort, node.socketTable.ConnPort, net.ParseIP(nodeCLI.DestIP), net.ParseIP(srcIP))
// 	conn.NodeSegSendChan = node.segSendChan
// 	conn.CLIChan = node.NodeCLIChan
// 	node.socketTable.OfferConn(conn)
// 	go conn.SynSend()
// 	return conn
// }

// *****************************************************************************************
// func (node *Node) HandleSendSegment(seg *proto.Segment) {
// 	hdr := seg.TCPhdr
// 	payload := seg.Payload
// 	tcpHeaderBytes := make(header.TCP, proto.TcpHeaderLen)
// 	tcpHeaderBytes.Encode(hdr)
// 	iPayload := make([]byte, 0, len(tcpHeaderBytes)+len(payload))
// 	iPayload = append(iPayload, tcpHeaderBytes...)
// 	iPayload = append(iPayload, []byte(payload)...)
// 	// proto.PrintHex(iPayload)
// 	node.RT.SendTCPPacket(seg.IPhdr.Src.String(), seg.IPhdr.Dst.String(), string(iPayload))
// }

// func (node *Node) handleRecvSegment(nodeCLI *proto.NodeCLI) {
// 	socketID := nodeCLI.Val16
// 	conn := node.socketTable.FindConnByID(socketID)
// 	if conn == nil {
// 		fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
// 		return
// 	}
// 	isBlock := nodeCLI.Bytes[0] == 'y'
// 	numBytes := nodeCLI.Val32
// 	conn.Retriv(numBytes, isBlock)
// }

// func (node *Node) handleClose(nodeCLI *proto.NodeCLI) {
// 	socketID := nodeCLI.Val16
// 	conn := node.socketTable.FindConnByID(socketID)
// 	if conn == nil {
// 		fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
// 		return
// 	}
// 	conn.Close()
// }

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
	conn.SendSegSYN()
	// seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn, conn.ISS), []byte{})
	// same thread => send in a goroutine
	// go node.SendToNodeSegChannel(seg)
	return conn, nil
}

// func (node *Node) SendToNodeSegChannel(segment *proto.Segment) {
// 	node.NodeSegSendChan <- segment
// }

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
	// fmt.Println(port, node.ST.Port2Listener)
	listener, ok := node.ST.Port2Listener(port)
	if ok {
		// fmt.Println("Sent to Listener")
		listener.ListenerSegRcvChan <- segment
		return
	}
	// socketID := nodeCLI.Val16
	// conn := node.socketTable.FindConnByID(socketID)
	// if conn == nil {
	// 	fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
	// 	return
	// }
	// isBlock := nodeCLI.Bytes[0] == 'y'
	// numBytes := nodeCLI.Val32
	// conn.Retriv(numBytes, isBlock)
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
	conn.WriteToSNDLoop(nodeCLI.Bytes)
	// conn.Write2SNDLoop
	// go conn.WriteIntoBuffer(nodeCLI.Bytes)
}
