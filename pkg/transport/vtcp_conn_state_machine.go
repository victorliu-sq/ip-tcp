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
		DPrintf("Conn [%v] Receive one segment with seqNum %v \n", conn.State, segment.TCPhdr.SeqNum)
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

// ***************************************************************************************
// 3WHS - Helper Functions
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

func (conn *VTCPConn) ToStateEstablished() {
	conn.State = proto.ESTABLISH
	go conn.SendSegmentLoop()
}

// ***************************************************************************************
// Established
func (conn *VTCPConn) HandleSegmentInStateESTABLISH(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// 1. Check if SeqNum is acceptable => rcv.NXT <= seqNum < rcv.NXT + rcv.WIN
	seqNum := segment.TCPhdr.SeqNum
	ackNum := segment.TCPhdr.AckNum
	segLen := len(segment.Payload)

	// Check flag SYN -> Retransmission for SND_RCVD
	if segment.TCPhdr.Flags&header.TCPFlagSyn != 0 {
		// 1.Send a 3WHS_ACK segment
		conn.SendSeg3WHS_ACK()
		DPrintf("[%v] gets one retransmitted SYN %v and sends one (ACK) segment in conn %v\n", conn.State, seqNum, conn.FormTuple())
		return
	}

	if segment.TCPhdr.Flags&header.TCPFlagAck != 0 {
		if len(segment.Payload) != 0 {
			// Check SeqNum
			if (conn.rcv.NXT <= seqNum && seqNum < conn.rcv.NXT+conn.rcv.WND) || (conn.rcv.NXT <= seqNum+uint32(segLen)-1 && seqNum+uint32(segLen)-1 < conn.rcv.NXT+conn.rcv.WND) {
				// go ahead
			} else {
				// DPrintf("Segment with SeqNum %v, Len %v gets Out of boundary", seqNum, segLen)
				return
			}
			DPrintf("---------------Receive one Segment with Payload ---------------")
			DPrintf("seqNum: %-16v\n", seqNum)
			DPrintf("payload: %-16v\n", string(segment.Payload))
			// write bytes into buffer
			isHeadAcked := conn.rcv.WriteSegmentToRCV(segment)
			if isHeadAcked {
				conn.rcond.Signal()
			}
			seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagAck, conn.snd.NXT, conn.rcv.NXT, conn.rcv.WND), []byte{})
			go conn.SendSeg(seg)
			conn.rcv.PrintRCV()
		} else {
			// Check AckNum
			if conn.snd.UNA >= ackNum {
				return
			}
			DPrintf("---------------Receive one Segment to ACK ---------------")
			conn.snd.SetUNA(ackNum)
			conn.snd.SetRCVWND(uint32(segment.TCPhdr.WindowSize))
			conn.wcv.Signal()
			if conn.snd.RCV_WND > 0 {
				conn.zeroProbe = false
				conn.scv.Signal()
			}
			conn.snd.PrintSND()
		}
	}
}
