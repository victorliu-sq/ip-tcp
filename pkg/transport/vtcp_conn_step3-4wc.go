package transport

import (
	"tcpip/pkg/proto"
	"time"

	"github.com/google/netstack/tcpip/header"
)

// ***************************************************************************************
// 4WC - Close Caller
func (conn *VTCPConn) Close() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if conn.State == proto.ESTABLISH {
		conn.ToStateFINWAIT1()
	} else if conn.State == proto.CLOSEWAIT {
		conn.ToStateLASTACK()
	}
}

func (conn *VTCPConn) WaitAndCLose() {
	conn.UpdateTimeout()
	time.Sleep(2 * proto.MSL)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if time.Now().After(conn.timeout2MSL) {
		conn.ToStateCLOSED()
	}
}

func (conn *VTCPConn) UpdateTimeout() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.timeout2MSL = time.Now().Add(2 * proto.MSL)
}

// ***************************************************************************************
// 4WC - Convert to FINWAIT1, FINWAIT2, TIMEWAIT, CLOSED
// ---------------------------- Client -------------------------------------
func (conn *VTCPConn) ToStateFINWAIT1() {
	// change state
	conn.State = proto.FINWAIT1
	// SendR Segment FIN1
	conn.SendSeg4WC_FIN()
}

func (conn *VTCPConn) ToStateFINWAIT2() {
	conn.State = proto.FINWAIT2
}

func (conn *VTCPConn) ToStateTIMEWAIT() {
	conn.State = proto.TIMEWAIT
}

func (conn *VTCPConn) ToStateCLOSED() {
	conn.State = proto.CLOSED
}

// ---------------------------- Server -------------------------------------
func (conn *VTCPConn) ToStateCLOSEWAIT() {
	// change state
	conn.State = proto.CLOSEWAIT
	// Send Segment ACK
	conn.SendSeg4WC_ACK()
	// Signal
	conn.rcond.Signal()
}

func (conn *VTCPConn) ToStateLASTACK() {
	// change state
	conn.State = proto.LASTACK
	// Send Segment ACK
	conn.SendSeg4WC_FIN()
}

// ***************************************************************************************
// 4WC - Sender
func (conn *VTCPConn) SendSeg4WC_FIN() {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagFin|header.TCPFlagAck, conn.snd.NXT, conn.rcv.NXT, conn.rcv.WND), []byte{})
	go conn.SendSegR(seg)
	DPrintf("[%v] Sends one (FIN|ACK) segment in conn %v\n", conn.State, conn.FormTuple())
}

func (conn *VTCPConn) SendSeg4WC_ACK() {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagAck, conn.snd.NXT, conn.rcv.NXT, conn.rcv.WND), []byte{})
	go conn.SendSeg(seg)
	DPrintf("[%v] Sends one (ACK) segment in conn %v\n", conn.State, conn.FormTuple())
}

// ***************************************************************************************
// 4WC - Handlers
func (conn *VTCPConn) HandleSegmentInStateFINWAIT1(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	seqNum := segment.TCPhdr.SeqNum
	ackNum := segment.TCPhdr.AckNum
	segLen := len(segment.Payload)

	// Check flag FIN => TIMEWAIT directly
	if segment.TCPhdr.Flags&header.TCPFlagFin != 0 {
		// 1.Send a 4WC_ACK segment
		conn.rcv.UpdateNXT_FIN()
		conn.ToStateTIMEWAIT()
		// update timeout2MSL, wait 2 MSL and try to close
		go conn.WaitAndCLose()
		conn.SendSeg4WC_ACK()
		DPrintf("[%v] gets one FIN %v and sends one (ACK) segment in conn %v\n", conn.State, seqNum, conn.FormTuple())
		return
	}

	if segment.TCPhdr.Flags&header.TCPFlagAck != 0 {
		if len(segment.Payload) != 0 {
			// 1. FIMWAIT1 can still receive segments
			// Check if SeqNum is acceptable => rcv.NXT <= seqNum < rcv.NXT + rcv.WIN
			if (conn.rcv.NXT <= seqNum && seqNum < conn.rcv.NXT+conn.rcv.WND) || (conn.rcv.NXT <= seqNum+uint32(segLen)-1 && seqNum+uint32(segLen)-1 < conn.rcv.NXT+conn.rcv.WND) {
				// go ahead
				DPrintf("---------------Receive one Segment with Payload ---------------")
				DPrintf("seqNum: %-16v\n", seqNum)
				DPrintf("payload: %-16v\n", string(segment.Payload))
				// write bytes into buffer
				isHeadAcked := conn.rcv.WriteSegmentToRCV(segment)
				if isHeadAcked {
					conn.rcond.Signal()
				}
				conn.rcv.PrintRCV()
			}
			conn.SendSegACK()
		} else {
			// 2. Check if ackNum is for FIN
			if conn.snd.UNA >= ackNum {
				return
			}
			DPrintf("---------------Receive one Segment to ACK ---------------")
			conn.snd.SetUNA(ackNum)
			conn.ToStateFINWAIT2()
		}
	}
}

func (conn *VTCPConn) HandleSegmentInStateFINWAIT2(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	seqNum := segment.TCPhdr.SeqNum
	// ackNum := segment.TCPhdr.AckNum
	segLen := len(segment.Payload)

	// Check flag FIN
	if segment.TCPhdr.Flags&header.TCPFlagFin != 0 {
		// 1.Send a 4WC_ACK segment
		conn.rcv.UpdateNXT_FIN()
		conn.ToStateTIMEWAIT()
		// update timeout2MSL, wait 2 MSL and try to close
		go conn.WaitAndCLose()
		conn.SendSeg4WC_ACK()
		DPrintf("[%v] gets one FIN %v and sends one (ACK) segment in conn %v\n", conn.State, seqNum, conn.FormTuple())
		return
	}

	if segment.TCPhdr.Flags&header.TCPFlagAck != 0 {
		if len(segment.Payload) != 0 {
			// 1. FIMWAIT2 can still receive segments
			// Check if SeqNum is acceptable => rcv.NXT <= seqNum < rcv.NXT + rcv.WIN
			if (conn.rcv.NXT <= seqNum && seqNum < conn.rcv.NXT+conn.rcv.WND) || (conn.rcv.NXT <= seqNum+uint32(segLen)-1 && seqNum+uint32(segLen)-1 < conn.rcv.NXT+conn.rcv.WND) {
				// go ahead
				DPrintf("---------------Receive one Segment with Payload ---------------")
				DPrintf("seqNum: %-16v\n", seqNum)
				DPrintf("payload: %-16v\n", string(segment.Payload))
				// write bytes into buffer
				isHeadAcked := conn.rcv.WriteSegmentToRCV(segment)
				if isHeadAcked {
					conn.rcond.Signal()
				}
				conn.rcv.PrintRCV()
			}
			conn.SendSegACK()
		}
	}
}

func (conn *VTCPConn) HandleSegmentInStateTIMEWAIT(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	seqNum := segment.TCPhdr.SeqNum
	// Check flag FIN
	if segment.TCPhdr.Flags&header.TCPFlagFin != 0 {
		// 1.Send a 4WC_ACK segment
		conn.rcv.UpdateNXT_FIN()
		// update timeout2MSL, wait 2 MSL and try to close
		go conn.WaitAndCLose()
		conn.SendSeg4WC_ACK()
		DPrintf("[%v] gets one FIN %v and sends one (ACK) segment in conn %v\n", conn.State, seqNum, conn.FormTuple())
		return
	}
}

// ---------------------------- Server -------------------------------------
func (conn *VTCPConn) HandleSegmentInStateCLOSEWAIT(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// seqNum := segment.TCPhdr.SeqNum
	ackNum := segment.TCPhdr.AckNum
	// segLen := len(segment.Payload)

	if segment.TCPhdr.Flags&header.TCPFlagAck != 0 {
		{
			// Client receives one segment
			// Check AckNum, update UNA and RCVWND, and signal writer to write into buffer
			if conn.snd.UNA >= ackNum {
				return
			}
			DPrintf("---------------Receive one Segment to ACK ---------------")
			conn.snd.SetUNA(ackNum)
			conn.snd.SetRCVWND(uint32(segment.TCPhdr.WindowSize))
			conn.wcond.Signal()
			if conn.snd.RCV_WND > 0 {
				conn.zeroProbe = false
				conn.scond.Signal()
			}
			conn.snd.PrintSND()
		}
	}
}

func (conn *VTCPConn) HandleSegmentInStateLASTACK(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// seqNum := segment.TCPhdr.SeqNum
	ackNum := segment.TCPhdr.AckNum
	// segLen := len(segment.Payload)
	if segment.TCPhdr.Flags&header.TCPFlagAck != 0 {
		{
			// Client receives one segment to ack FIN
			if conn.snd.UNA >= ackNum {
				return
			}
			DPrintf("---------------Receive one Segment to ACK ---------------")
			conn.snd.SetUNA(ackNum)
			conn.ToStateCLOSED()
		}
	}
}
