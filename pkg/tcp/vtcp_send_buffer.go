package tcp

import (
	"fmt"
	"tcpip/pkg/proto"
)

type SendBuffer struct {
	buffer []byte
	isn    uint32 // initial sequence number
	una    uint32 // oldest unacked byte
	nxt    uint32 // next byte to send
	lbw    uint32 // last bytes written
	//
	total uint32 // total number of bytes in buffer
	win   uint32 // window size in send buffer
}

func NewSendBuffer(seqNum, winSize uint32) *SendBuffer {
	sb := &SendBuffer{
		buffer: make([]byte, proto.BUFFER_SIZE),
		isn:    seqNum,
		una:    seqNum,
		nxt:    seqNum,
		lbw:    seqNum,
		total:  0,
		win:    winSize,
	}
	// fmt.Printf("window size is %v\n", winSize)
	return sb
}

// *********************************************************************************************
// Write bytes into send buffer
func (sb *SendBuffer) WriteIntoBuffer(content []byte) uint32 {
	remain := sb.GetRemainBytes()
	// fmt.Printf("Remaining space is %v bytes\n", remain)
	bnum := uint32(len(content))
	// 1. if not enough space, only write part of content into buffer
	if remain < uint32(len(content)) {
		content = content[:remain]
		bnum = remain
	}
	// 2. write bytes of content into end as many as possible
	remainBack := sb.getRemainingBytesBack()
	// 2.(1) if all bytes can be written, write once
	if remainBack > uint32(len(content)) {
		copy(sb.buffer[sb.getIdx(sb.lbw):], content)
	} else {
		// 2.(2) Otherwise, write twice
		// <1> write part of content into right part of buffer
		copy(sb.buffer[sb.getIdx(sb.lbw):], content[:remainBack])
		content2 := content[remainBack:]
		// <2> write remaining part of content into left of buffer
		copy(sb.buffer, content2)
	}
	// 3. update total
	sb.lbw += bnum
	sb.total += bnum
	return bnum
}

// *********************************************************************************************
// Send out one segment
// Check if there are bytes left in the buffer to send
func (sb *SendBuffer) CanSend() bool {
	return sb.nxt < sb.lbw
}

func (sb *SendBuffer) GetSegmentToSendAndUpdateNxt(mtu int) ([]byte, uint32) {
	var len uint32
	seqNum := sb.nxt
	if sb.win < uint32(mtu) {
		mtu = int(sb.win)
	}

	if sb.nxt+uint32(mtu) > sb.lbw {
		len = sb.lbw - sb.nxt
		// payload = sb.buffer[sb.getIdx(sb.nxt):sb.getIdx(sb.lbw)]
	} else {
		len = uint32(mtu)
		// payload = sb.buffer[sb.getIdx(sb.nxt):sb.getIdx(sb.nxt+uint32(mtu))]
	}

	payload := make([]byte, len)
	if sb.getIdx(sb.nxt)+len < proto.BUFFER_SIZE {
		// copy all, notice that there must be <, <= will cause range problem like [9:0]
		copy(payload, sb.buffer[sb.getIdx(sb.nxt):sb.getIdx(sb.nxt+len)])
	} else {
		// copy right and left
		len1 := proto.BUFFER_SIZE - sb.getIdx(sb.nxt)
		len2 := len - len1
		copy(payload, sb.buffer[sb.getIdx(sb.nxt):proto.BUFFER_SIZE])
		copy(payload[len1:], sb.buffer[:len2])
	}
	// Update metadata of send buffer
	if sb.win != 0 {
		sb.nxt += len
		sb.win -= len
	}
	return payload, seqNum
}

func (sb *SendBuffer) GetZeroProbe() ([]byte, uint32) {
	seqNum := sb.nxt
	len := uint32(1)
	payload := make([]byte, len)
	copy(payload, sb.buffer[sb.getIdx(sb.nxt):sb.getIdx(sb.nxt)+len])
	sb.nxt += 1
	return payload, seqNum
}

// 12345 a bcde//
// isZeroProbeSend = false
// send zeroProbe => isZeroProbesent = true
// *********************************************************************************************
// Receive out one ACK
func (sb *SendBuffer) UpdateUNA(ack *proto.Segment) uint32 {
	acked := uint32(0)
	if ack.TCPhdr.AckNum > sb.una {
		fmt.Println(ack.TCPhdr.AckNum)
		// length of payload is (ackNum - sb.una)
		for curSeqNum := sb.una; curSeqNum < ack.TCPhdr.AckNum; curSeqNum++ {
			sb.buffer[sb.getIdx(curSeqNum)] = byte('*')
			acked += 1
		}
		sb.total -= acked
		sb.una += acked
	}
	return acked
}

// *********************************************************************************************
// Helper function
func (sb *SendBuffer) GetRemainBytes() uint32 {
	return proto.BUFFER_SIZE - sb.total
}

func (sb *SendBuffer) getRemainingBytesBack() uint32 {
	return proto.BUFFER_SIZE - 1 - sb.getIdx(sb.lbw) + 1
}

func (sb *SendBuffer) getIdx(seqNum uint32) uint32 {
	return (seqNum - sb.isn) % proto.BUFFER_SIZE
}

func (sb *SendBuffer) UpdateWin(tcpHeaderWin uint16) {
	sb.win = uint32(tcpHeaderWin)
}

// Check if current send buffer is full
func (sb *SendBuffer) IsFull() bool {
	return sb.total == proto.BUFFER_SIZE
}
