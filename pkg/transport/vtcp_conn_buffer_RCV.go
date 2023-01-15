package transport

import (
	"fmt"
	"tcpip/pkg/proto"
)

type RCV struct {
	buffer []byte
	ISS    uint32
	LBR    uint32
	NXT    uint32
	WND    uint32
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

func (rcv *RCV) WriteSegmentToRCV(segment *proto.Segment) (uint32, bool) {
	// 1. Get number of bytes to write int buffer
	bnum := getMinLengthRCV(rcv.getRemainingBytes(), uint32(len(segment.Payload)))
	isHead := segment.TCPhdr.SeqNum == rcv.NXT
	// 2. write bytes of Payload into RCV as many as possible
	content := segment.Payload
	remainR := rcv.getRemainingBytesRight()
	// 2.(1) if all bytes can be written into right part, write once
	if remainR > uint32(len(content)) {
		copy(rcv.buffer[rcv.getIdx(rcv.NXT):], content)
	} else {
		// 2.(2) Otherwise, write twice
		// <1> write remainR bytes of content into right part of buffer
		copy(rcv.buffer[rcv.getIdx(rcv.NXT):], content[:remainR])
		content2 := content[remainR:]
		// <2> write remainL part of content into left of buffer
		copy(rcv.buffer, content2)
	}
	// 3. update total and LBW
	if isHead {
		rcv.NXT += bnum
		rcv.total += bnum
	}
	return bnum, isHead
}

// *********************************************************************************************
// Helper function
func getMinLengthRCV(remainBytes, segmentLen uint32) uint32 {
	if remainBytes < segmentLen {
		return remainBytes
	}
	return segmentLen
}

func (rcv *RCV) getIdx(seqNum uint32) uint32 {
	return (seqNum - rcv.ISS - 1) % proto.RCV_BUFFER_SIZE
}

func (rcv *RCV) getRemainingBytes() uint32 {
	return proto.RCV_BUFFER_SIZE - rcv.total
}

func (rcv *RCV) getRemainingBytesRight() uint32 {
	return proto.RCV_BUFFER_SIZE - 1 - rcv.getIdx(rcv.NXT) + 1
}
