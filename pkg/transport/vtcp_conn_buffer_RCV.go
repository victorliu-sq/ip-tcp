package transport

import (
	"fmt"
	"tcpip/pkg/proto"
)

type RCV struct {
	buffer []byte
	ISS    uint32
	LBR    uint32
	NXT    uint32 // UNA in SND
	WND    uint32 // RCV_WND in SND
	total  uint32
}

func NewRCV(ClientISS uint32) *RCV {
	rcv := &RCV{
		buffer: make([]byte, proto.RCV_BUFFER_SIZE),
		ISS:    ClientISS,
		NXT:    ClientISS + 1,
		LBR:    ClientISS + 1,
		WND:    proto.RCV_BUFFER_SIZE,
		total:  0,
	}
	for i := 0; i < int(proto.RCV_BUFFER_SIZE); i++ {
		rcv.buffer[i] = byte('*')
	}
	return rcv
}

func (rcv *RCV) PrintRCV() {
	DPrintf("----------------RCV----------------\n")
	DPrintf("%-16v %-16v %-16v %-16v\n", "ISS", "NXT", "LBR", "WIN")
	DPrintf("%-16v %-16v %-16v %-16v\n", rcv.ISS, rcv.NXT, rcv.LBR, rcv.WND)
	DPrintf("RCV buffer: %v\n", fmt.Sprintf("%v", string(rcv.buffer)))
}

func (rcv *RCV) WriteSegmentToRCV(segment *proto.Segment) bool {
	payload := segment.Payload
	seqNum := segment.TCPhdr.SeqNum
	isHeadAcked := false
	ackedNum := 0
	for i, ch := range payload {
		curSeqNum := seqNum + uint32(i)
		if curSeqNum == rcv.NXT {
			isHeadAcked = true
		}
		if rcv.NXT <= curSeqNum && curSeqNum < rcv.NXT+proto.RCV_BUFFER_SIZE {
			idx := rcv.getIdx(seqNum + uint32(i))
			rcv.buffer[idx] = ch
			ackedNum += 1
		}
	}
	if isHeadAcked {
		rcv.NXT += uint32(ackedNum)
		rcv.total += uint32(ackedNum)
		rcv.WND -= uint32(ackedNum)
	}
	return isHeadAcked
}

func (rcv *RCV) ReadFromBuffer(total uint32) ([]byte, uint32) {
	bytes := []byte{}
	bnum := uint32(0)
	for bnum < total && rcv.LBR < rcv.NXT {
		// Reset buffer
		idx := rcv.getIdx(rcv.LBR)
		bytes = append(bytes, rcv.buffer[idx])
		rcv.buffer[idx] = byte('*')

		// Update metadata
		bnum += 1
		rcv.LBR += 1
		rcv.WND += 1
		rcv.total -= 1
	}
	return bytes, bnum
}

// *********************************************************************************************
// Helper function
func (rcv *RCV) getIdx(seqNum uint32) uint32 {
	return (seqNum - rcv.ISS - 1) % proto.RCV_BUFFER_SIZE
}

func (rcv *RCV) IsEmpty() bool {
	return rcv.total == 0
}
