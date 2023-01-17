package transport

import (
	"tcpip/pkg/proto"
	"time"
)

// ***************************************************************************************
// retransmission
func (conn *VTCPConn) RetransmissionLoop() {
	for {
		segmentR := <-conn.retsmChan
		go conn.Retransmit(segmentR)
	}
}

func (conn *VTCPConn) Retransmit(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	seqNum := segment.TCPhdr.SeqNum
	if conn.snd.UNA > seqNum {
		// Check if segment has been acked
		DPrintf("Conn %v has successfully acked retransmitted segment with seqNum %v\n", conn.ID, seqNum)
		// DPrintf("Segment is %v\n", segment)
		return
	}
	// if not acked, retransmit it again
	time.Sleep(proto.RetranInterval)
	// Sleep later
	DPrintf("Conn %v retransmits an unacked segment with seqNum %v, SND.UNA is %v \n", conn.ID, seqNum, conn.snd.UNA)
	go conn.SendSegR(segment)
}
