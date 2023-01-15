package transport

import (
	"fmt"
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

// Normal socket State machine
func (conn *VTCPConn) ConnStateMachineLoop() {
	for {
		segment := <-conn.ConnSegRcvChan
		// fmt.Println(segment)
		DPrintf("Conn [%v] %v Receive one segment with seqNum %v \n", conn.State, conn.FormTuple(), segment.TCPhdr.SeqNum)
		switch conn.State {
		case proto.SYN_SENT:
			conn.HandleSegmentInStateSYNSENT(segment)
		case proto.SYN_RCVD:
			conn.HandleSegmentInStateSYNRCVD(segment)
		case proto.ESTABLISH:
			conn.HandleSegmentInStateESTABLISH(segment)
		}
	}
}

// ***************************************************************************************
// 3WHS - Handlers

func (conn *VTCPConn) HandleSegmentInStateSYNSENT(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// 1. Check flag ACK
	if segment.TCPhdr.Flags|header.TCPFlagAck != 0 {
		// Advance SND.UNA
		if conn.snd.UNA >= segment.TCPhdr.AckNum {
			return
		}
		conn.snd.AdvanceUNA(segment.TCPhdr.AckNum)
		DPrintf("[%v] SND.UNA is advanced to %v\n", conn.State, conn.snd.UNA)
	}
	// 2. Check flag SYN
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
		DPrintf("[%v] conn %v converts to state Established\n", conn.State, conn.ID)
		conn.snd.PrintSND()
		conn.rcv.PrintRCV()
	}
}

func (conn *VTCPConn) HandleSegmentInStateSYNRCVD(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// Check flag ACK
	if segment.TCPhdr.Flags|header.TCPFlagAck != 0 {
		// Advance SND.UNA
		if conn.snd.UNA >= segment.TCPhdr.AckNum {
			return
		}
		conn.snd.AdvanceUNA(segment.TCPhdr.AckNum)
		DPrintf("[%v] SND.UNA is advanced to %v\n", conn.State, conn.snd.UNA)
		// Covert to State Established
		DPrintf("[%v] conn %v converts to state Established\n", conn.State, conn.ID)
		conn.ToStateEstablished()
		conn.rcv.PrintRCV()
		conn.snd.PrintSND()
	}
}

// ***************************************************************************************
// 3WHS - Helper Functions
func (conn *VTCPConn) SendSegSYN() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn, conn.ISN, 0, 0), []byte{})
	go conn.SendSegR(segment)
	DPrintf("[%v] Sends one (SYN) segment in conn %v\n", conn.State, conn.FormTuple())
}

func (conn *VTCPConn) SendSegSYNACK() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.ISN, conn.rcv.NXT, conn.rcv.WND), []byte{})
	go conn.SendSegR(segment)
	DPrintf("[%v] Sends one (SYN | ACK) segment in conn %v\n", conn.State, conn.FormTuple())
}

func (conn *VTCPConn) SendSegACK() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, 0, conn.rcv.NXT, conn.rcv.WND), []byte{})
	go conn.SendSeg(segment)
	DPrintf("[%v] Sends one (ACK) segment in conn %v\n", conn.State, conn.FormTuple())
}

func (conn *VTCPConn) ToStateEstablished() {
	conn.State = proto.ESTABLISH
	go conn.SendSegmentLoop()
}

// ***************************************************************************************
// Established
func (conn *VTCPConn) HandleSegmentInStateESTABLISH(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// Check flag SYN -> Retransmission for SND_RCVD
	if segment.TCPhdr.Flags|header.TCPFlagSyn != 0 {
		// 1.Send a ACK segment
		conn.SendSegACK()
		fmt.Println(segment.TCPhdr.SeqNum)
		DPrintf("[%v] Sends one (ACK) segment in conn %v\n", conn.State, conn.FormTuple())
	}
	// Check SeqNum
}
