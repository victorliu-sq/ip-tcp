package transport

import (
	"tcpip/pkg/proto"
	"time"
)

// ***************************************************************************************
// retransmission
func (conn *VTCPConn) RetransmissionLoop() {
	for {
		segmentR := <-conn.rtmQueue
		go conn.Retransmit(segmentR)
	}
}

func (conn *VTCPConn) Retransmit(segment *proto.Segment) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	time.Sleep(proto.RetranInterval)
	seqNum := segment.TCPhdr.SeqNum
	if conn.snd.CheckACK(seqNum) {
		// Check if segment has been acked
		DPrintf("Conn %v has successfully acked retransmitted segment with seqNum %v\n", conn.ID, seqNum)
		// DPrintf("Segment is %v\n", segment)
		return
	}
	// if not acked, retransmit it again
	DPrintf("Conn %v has retransmits unacked segment with seqNum %v, SND.UNA is %v \n", conn.ID, seqNum, conn.snd.UNA)
	conn.SendSegR(segment)
}
