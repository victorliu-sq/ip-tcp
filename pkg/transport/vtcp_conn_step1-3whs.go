package transport

import (
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

// ***************************************************************************************
// 3WHS - Convert to State
func (conn *VTCPConn) ToStateEstablished() {
	conn.State = proto.ESTABLISH
	go conn.SendSegmentLoop()
}

// ***************************************************************************************
// 3WHS - Sender
func (conn *VTCPConn) SendSeg3WHS_SYN() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn, conn.ISN, 0, 0), []byte{})
	go conn.SendSegR(segment)
	DPrintf("[%v] Sends one (SYN) segment in conn %v\n", conn.State, conn.FormTuple())
}

func (conn *VTCPConn) SendSeg3WHS_SYNACK() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.ISN, conn.rcv.NXT, conn.rcv.WND), []byte{})
	go conn.SendSegR(segment)
	DPrintf("[%v] Sends one (SYN | ACK) segment in conn %v\n", conn.State, conn.FormTuple())
}

func (conn *VTCPConn) SendSeg3WHS_ACK() {
	segment := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.snd.NXT, conn.rcv.NXT, conn.rcv.WND), []byte{})
	go conn.SendSeg(segment)
	DPrintf("[%v] Sends one (ACK) segment in conn %v\n", conn.State, conn.FormTuple())
}

// ***************************************************************************************
// 3WHS - Handlers

func (conn *VTCPConn) HandleSegmentInStateSYNSENT(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// 1. Check flag ACK
	if segment.TCPhdr.Flags&header.TCPFlagAck != 0 {
		// Advance SND.UNA
		if conn.snd.UNA >= segment.TCPhdr.AckNum {
			return
		}
		conn.snd.SetUNA(segment.TCPhdr.AckNum)
		conn.snd.SetRCVWND(uint32(segment.TCPhdr.WindowSize))
		DPrintf("[%v] SND.UNA is advanced to %v\n", conn.State, conn.snd.UNA)
	}
	// 2. Check flag SYN
	if segment.TCPhdr.Flags&header.TCPFlagSyn != 0 {
		// 1. Synchronize SeqNum
		// Create a RCV for conn and Set RCV.NXT = seqNum + 1
		conn.rcv = NewRCV(segment.TCPhdr.SeqNum)
		conn.rcv.NXT = segment.TCPhdr.SeqNum + 1
		DPrintf("[%v] RCV.NXT is initialized to %v\n", conn.State, conn.rcv.NXT)
		// 2 Send a ACK segment
		conn.SendSeg3WHS_ACK()
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
	if segment.TCPhdr.Flags&header.TCPFlagAck != 0 {
		// Advance SND.UNA
		if conn.snd.UNA >= segment.TCPhdr.AckNum {
			return
		}
		conn.snd.SetUNA(segment.TCPhdr.AckNum)
		conn.snd.SetRCVWND(uint32(segment.TCPhdr.WindowSize))
		DPrintf("[%v] SND.UNA is advanced to %v\n", conn.State, conn.snd.UNA)
		// Covert to State Established
		DPrintf("[%v]  converts to state Established\n", conn.State)
		conn.ToStateEstablished()
		conn.rcv.PrintRCV()
		conn.snd.PrintSND()
	}
}
