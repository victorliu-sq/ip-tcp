package kernel

import (
	"fmt"
	"net"
	"strconv"
	"tcpip/pkg/proto"
	"tcpip/pkg/tcp"

	"github.com/google/netstack/tcpip/header"
)

func (node *Node) handleTCPSegment() {
	for {
		segment := <-node.segRecvChan
		tuple := segment.FormTuple()
		if conn := node.socketTable.FindConn(tuple); conn != nil {
			conn.SegRcvChan <- segment
			continue
		}
		dstPort := segment.TCPhdr.DstPort
		listener := node.socketTable.FindListener(dstPort)
		if listener != nil {
			listener.SegRcvChan <- segment
		}
	}
}

// *****************************************************************************************
// Handle Create Listener
func (node *Node) handleCreateListener(msg *proto.NodeCLI) {
	val, _ := strconv.Atoi(msg.Msg)
	port := uint16(val)
	if node.socketTable.FindListener(port) != nil {
		fmt.Printf("Cannot assign requested address\n")
	} else {
		listener := node.socketTable.OfferListener(port)
		go node.NodeAcceptLoop(listener, false)
	}
}

// a port -> listener  -> go node.acceptConn(listener)
func (node *Node) NodeAcceptLoop(listener *tcp.VTCPListener, oneTime bool) {
	for {
		conn, err := listener.VAccept()
		if err != nil {
			continue
		}
		conn.NodeSegSendChan = node.segSendChan
		conn.CLIChan = node.NodeCLIChan
		conn.RcvBuf = tcp.NewRecvBuffer(conn.GetAck(), tcp.DEFAULTWINDOWSIZE)
		node.socketTable.OfferConn(conn)
		// Recv SYN
		go conn.SynRev()
		if oneTime {
			listener.CancelChan <- true
			for listener.ConnInQueue > 0 {
				conn := <-listener.ConnQueue
				node.socketTable.DeleteSocket(conn.ID)
			}
			cli := <-listener.CLIChan
			node.socketTable.DeleteSocket(listener.ID)
			go conn.RetrivFile(cli.Fd)
		}
	}
}

// *****************************************************************************************
// Handle Create Conn
func (node *Node) HandleCreateConn(nodeCLI *proto.NodeCLI) *tcp.VTCPConn {
	// Create a Normal Socket
	srcIP := node.RT.FindSrcIPAddr(nodeCLI.DestIP)
	if srcIP == "no" {
		fmt.Println("v_connect() error: No route to host")
		return nil
	}
	conn := tcp.NewNormalSocket(0, nodeCLI.DestPort, node.socketTable.ConnPort, net.ParseIP(nodeCLI.DestIP), net.ParseIP(srcIP))
	conn.NodeSegSendChan = node.segSendChan
	conn.CLIChan = node.NodeCLIChan
	node.socketTable.OfferConn(conn)
	go conn.SynSend()
	return conn
}

// *****************************************************************************************
// Handle Send Bytes
func (node *Node) handleSendSegment(nodeCLI *proto.NodeCLI) {
	socketID := nodeCLI.Val16
	conn := node.socketTable.FindConnByID(socketID)
	if conn == nil {
		fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
		return
	}
	go conn.VSBufferWrite(nodeCLI.Bytes)
}

// *****************************************************************************************
func (node *Node) HandleSendOutSegment(seg *proto.Segment) {
	hdr := seg.TCPhdr
	payload := seg.Payload
	tcpHeaderBytes := make(header.TCP, proto.TcpHeaderLen)
	tcpHeaderBytes.Encode(hdr)
	iPayload := make([]byte, 0, len(tcpHeaderBytes)+len(payload))
	iPayload = append(iPayload, tcpHeaderBytes...)
	iPayload = append(iPayload, []byte(payload)...)
	// proto.PrintHex(iPayload)
	node.RT.SendTCPPacket(seg.IPhdr.Src.String(), seg.IPhdr.Dst.String(), string(iPayload))
}

func (node *Node) handleRecvSegment(nodeCLI *proto.NodeCLI) {
	socketID := nodeCLI.Val16
	conn := node.socketTable.FindConnByID(socketID)
	if conn == nil {
		fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
		return
	}
	isBlock := nodeCLI.Bytes[0] == 'y'
	numBytes := nodeCLI.Val32
	conn.Retriv(numBytes, isBlock)
}

func (node *Node) handleClose(nodeCLI *proto.NodeCLI) {
	socketID := nodeCLI.Val16
	conn := node.socketTable.FindConnByID(socketID)
	if conn == nil {
		fmt.Printf("no VTCPConn with socket ID %v\n", socketID)
		return
	}
	conn.Close()
}
