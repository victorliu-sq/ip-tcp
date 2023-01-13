package transport

import (
	"tcpip/pkg/proto"
	"time"
)

// ***************************************************************************************
// retransmission
func (conn *VTCPConn) retransmit(segment *proto.Segment) {
	time.Sleep(proto.RetranInterval)
	seqNum := segment.TCPhdr.SeqNum
	if conn.snd.CheckACK(seqNum) {
		// Check if segment has been acked
		DPrintf("Conn %v has successfully acked segment with seqNum %v\n", conn.ID, seqNum)
		// DPrintf("Segment is %v\n", segment)
		return
	}
	// if not acked, retransmit it again
	DPrintf("Conn %v has retransmits unacked segment with seqNum %v, SND.UNA is %v \n", conn.ID, seqNum, conn.snd.UNA)
	go conn.SendSegR(segment)
}
