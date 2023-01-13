package transport

import (
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

// Normal socket State machine
func (conn *VTCPConn) ConnStateMachineLoop() {
	for {
		segment := <-conn.ConnSegRcvChan
		// fmt.Println(segment)
		DPrintf("Conn [%v] %v Receive one segment \n", conn.State, conn.FormTuple())
		switch conn.State {
		case proto.SYN_SENT:
			conn.HandleSegmentInStateSYNSENT(segment)
		case proto.SYN_RCVD:
			conn.HandleSegmentInStateSYNRCVD(segment)
		}
	}
}

func (conn *VTCPConn) ToStateEstablished() {
	conn.State = proto.ESTABLISH
}

// ***************************************************************************************
// 3WHS

func (conn *VTCPConn) HandleSegmentInStateSYNSENT(segment *proto.Segment) {
	// Check flag ACK
	if segment.TCPhdr.Flags|header.TCPFlagAck != 0 {
		// Advance SND.UNA
		conn.snd.AdvanceUNA(segment.TCPhdr.AckNum)
		DPrintf("[%v] SND.UNA is advanced to %v\n", conn.State, conn.snd.UNA)
	}
	// Check flag SYN
	if segment.TCPhdr.Flags|header.TCPFlagSyn != 0 {
		// 1. Synchronize SeqNum
		// Create a RCV for conn and Set RCV.NXT = seqNum + 1
		conn.rcv = NewRCV(segment.TCPhdr.SeqNum)
		conn.rcv.NXT = segment.TCPhdr.SeqNum + 1
		DPrintf("[%v] RCV.NXT is initialized to %v\n", conn.State, conn.rcv.NXT)
		// 2 Send a ACK segment
		conn.SendSegACK()
		DPrintf("[%v] Sends one (ACK) segment in conn %v\n", conn.State, conn.FormTuple())
		// 3. Enter State Established
		conn.ToStateEstablished()
		conn.snd.PrintSND()
		conn.rcv.PrintRCV()
	}
}

func (conn *VTCPConn) HandleSegmentInStateSYNRCVD(segment *proto.Segment) {
	// Check flag ACK
	if segment.TCPhdr.Flags|header.TCPFlagSyn != 0 {
		// Advance SND.UNA
		conn.snd.AdvanceUNA(segment.TCPhdr.AckNum)
		DPrintf("[%v] SND.UNA is advanced to %v\n", conn.State, conn.snd.UNA)
		// Covert to State Established
		conn.ToStateEstablished()
		conn.rcv.PrintRCV()
		conn.snd.PrintSND()
	}
}

func (conn *VTCPConn) SendSegSYN() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn, conn.ISN, 0), []byte{})
	go conn.SendSeg(segment)
	DPrintf("[%v] Sends one (SYN) segment in conn %v\n", conn.State, conn.FormTuple())
}

func (conn *VTCPConn) SendSegSYNACK() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.ISN, conn.rcv.NXT), []byte{})
	go conn.SendSeg(segment)
	DPrintf("[%v] Sends one (SYN | ACK) segment in conn %v\n", conn.State, conn.FormTuple())

}

func (conn *VTCPConn) SendSegACK() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, 0, conn.rcv.NXT), []byte{})
	go conn.SendSeg(segment)
}

func (conn *VTCPConn) SendSeg(segment *proto.Segment) {
	conn.NodeSegSendChan <- segment
}

// ***************************************************************************************
// retransmission
