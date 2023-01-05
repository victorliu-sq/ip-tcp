package transport

import (
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

func (conn *VTCPConn) ConnStateMachine() {
	for {
		segment := <-conn.ConnSegRcvChan
		// fmt.Println(segment)
		switch conn.State {
		case proto.SYN_SENT:
			DPrintf("[3WHS-Client] conn %v Receive one segment in SYN_SENT \n", conn.FormTuple())
			conn.HandleSegmentSYNSENT(segment)
		case proto.SYN_RCVD:
			DPrintf("[3WHS-Server] conn %v Receive one segment in SYN_RCVD\n", conn.FormTuple())
			conn.HandleSegmentSYNRCVD(segment)
		}
	}
}

func (conn *VTCPConn) SendSegSYN() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn, conn.ISS), []byte{})
	conn.NodeSegSendChan <- segment
	DPrintf("[3WHS-Client] Sends one SYN segment in conn %v\n", conn.FormTuple())
}

func (conn *VTCPConn) SendSegSYNACK() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.ISS), []byte{})
	conn.NodeSegSendChan <- segment
	DPrintf("[3WHS-Server] Sends one SYN | ACK segment in conn %v\n", conn.FormTuple())
}

func (conn *VTCPConn) HandleSegmentSYNSENT(segment *proto.Segment) {
}

func (conn *VTCPConn) HandleSegmentSYNRCVD(segment *proto.Segment) {
}
