package transport

import (
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

// Write bytes into SND Buffer
func (conn *VTCPConn) Write2SNDLoop(content []byte) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	total := uint32(len(content))
	for total > 0 {
		if !conn.snd.IsFull() {
			bnum := conn.snd.WriteIntoBuffer(content)
			total -= bnum
			content = content[bnum:]
			conn.scv.Signal()
			// Print current snd
			conn.snd.PrintSND()
		} else {
			conn.wcv.Wait()
		}
	}
}

// Send Segment from bytes in SND Buffer
func (conn *VTCPConn) SendSegmentLoop() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	for conn.State == proto.ESTABLISH {
		if conn.snd.CanSend() && !conn.zeroProbe {
			var payload []byte
			var seqNum uint32
			if conn.snd.WND == 0 {
				payload, seqNum = conn.snd.GetZeroProbe()
				seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagAck, seqNum, conn.snd.UNA, 0), payload)
				go conn.SendSegR(seg)
				// conn.send(payload, seqNum)
				conn.zeroProbe = true
			} else {
				// Get one segment, send it out and add it to retransmission queue
				payload, seqNum = conn.snd.GetSegment()
				seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.BuildTCPHdr(header.TCPFlagAck, seqNum, conn.snd.UNA, 0), payload)
				go conn.SendSegR(seg)
				// conn.send(payload, seqNum)
			}
			DPrintf("---------------Send one Segment---------------")
			DPrintf("%-16v %-16v\n", "seqNum", "PayLoad")
			DPrintf("%-16v %-16v\n", seqNum, string(payload))
		} else {
			conn.scv.Wait()
		}
	}
}

func (conn *VTCPConn) SendSeg(segment *proto.Segment) {
	conn.NodeSegSendChan <- segment
}

// send a segment that will be retransmitted
func (conn *VTCPConn) SendSegR(segment *proto.Segment) {
	conn.NodeSegSendChan <- segment
	conn.rtmQueue <- segment
}
