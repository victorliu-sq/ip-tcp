package transport

import (
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

// ***************************************************************************************
// Established
// Receiver
// Segment Handler => Write Segment 2 Receive + Send ACK / Set UNA by ACK
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

	// Check flag FIN
	if segment.TCPhdr.Flags&header.TCPFlagFin != 0 {
		// 1.Send a 4WC_ACK segment
		conn.rcv.UpdateNXT_FIN()
		conn.SendSeg4WC_ACK()
		conn.ToStateCLOSEWAIT()
		DPrintf("[%v] gets one FIN %v and sends one (ACK) segment in conn %v\n", conn.State, seqNum, conn.FormTuple())
		return
	}

	// Server receives one segment
	if segment.TCPhdr.Flags&header.TCPFlagAck != 0 {
		if len(segment.Payload) != 0 {
			// Check SeqNum, write segment into buffer, signal reader to read from buffer if header is acked
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

// ***************************************************************************************
// Established - Sender
// Send Segment from bytes in SND Buffer
func (conn *VTCPConn) SegmentSender() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	for conn.State == proto.ESTABLISH || conn.State == proto.CLOSEWAIT {
		if conn.snd.CanSend() && !conn.zeroProbe {
			var payload []byte
			var seqNum uint32
			if conn.snd.RCV_WND == 0 {
				payload, seqNum = conn.snd.ReadZeroProbeFromSND()
				seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagAck, seqNum, conn.rcv.NXT, conn.rcv.WND), payload)
				go conn.SendSegR(seg)
				// conn.send(payload, seqNum)
				conn.zeroProbe = true
			} else {
				// Get one segment, send it out and add it to retransmission queue
				payload, seqNum = conn.snd.ReadSegmentFromSND()
				seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagAck, seqNum, conn.rcv.NXT, conn.rcv.WND), payload)
				go conn.SendSegR(seg)
				// conn.send(payload, seqNum)
			}
			DPrintf("---------------Send one Segment---------------")
			DPrintf("%-16v %-16v\n", "seqNum", "PayLoad")
			DPrintf("%-16v %-16v\n", seqNum, string(payload))
		} else {
			conn.scond.Wait()
		}
	}
}

// ***************************************************************************************
func (conn *VTCPConn) SendSegACK() {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagAck, conn.snd.NXT, conn.rcv.NXT, conn.rcv.WND), []byte{})
	go conn.SendSeg(seg)
	// DPrintf("[%v] Sends one (SYN) segment in conn %v\n", conn.State, conn.FormTuple())
}

func (conn *VTCPConn) SendSeg(segment *proto.Segment) {
	conn.NodeSegSendChan <- segment
}

// send a segment that will be retransmitted
func (conn *VTCPConn) SendSegR(segment *proto.Segment) {
	conn.NodeSegSendChan <- segment
	conn.retsmChan <- segment
}
