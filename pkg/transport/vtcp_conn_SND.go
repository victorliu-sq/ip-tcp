package transport

import (
	"tcpip/pkg/proto"
)

type SND struct {
	buffer []byte
	ISS    uint32
	UNA    uint32
	NXT    uint32
	LBW    uint32
	// metadata
	WND   uint32
	total uint32
}

func NewSND(ISN uint32) *SND {
	snd := &SND{
		buffer: make([]byte, proto.SND_BUFFER_SIZE),
		ISS:    ISN,
		UNA:    ISN,
		NXT:    ISN + 1,
		LBW:    ISN + 1,
		// initialize total as 1
		WND:   proto.SND_BUFFER_SIZE,
		total: 1,
	}
	for i := 0; i < int(proto.SND_BUFFER_SIZE); i++ {
		snd.buffer[i] = byte('*')
	}
	return snd
}

func (snd *SND) PrintSND() {
	DPrintf("----------------SND----------------\n")
	DPrintf("%-16v %-16v %-16v %-16v %-16v\n", "ISS", "UNA", "NXT", "LBW", "WIN")
	DPrintf("%-16v %-16v %-16v %-16v %-16v\n", snd.ISS, snd.UNA, snd.NXT, snd.LBW, snd.WND)
	DPrintf("SND buffer: %v\n", string(snd.buffer))
}

func (snd *SND) AdvanceUNA(ackNum uint32) {
	snd.total -= (ackNum - 1 - snd.UNA + 1)
	snd.UNA = ackNum
}

func (snd *SND) CheckACK(seqNum uint32) bool {
	return snd.UNA > seqNum
}

// *********************************************************************************************
// LBW => Write bytes into send buffer
func (snd *SND) WriteIntoBuffer(content []byte) uint32 {
	remain := snd.getRemainingBytes()
	// fmt.Printf("Remaining space is %v bytes\n", remain)
	bnum := uint32(len(content))
	// 1. if not enough space, only write part of content into buffer
	if remain < uint32(len(content)) {
		content = content[:remain]
		bnum = remain
	}
	// 2. write bytes of content into right part of send buffer as many as possible
	remainR := snd.getRemainingBytesRight()
	// 2.(1) if all bytes can be written into right part, write once
	if remainR > uint32(len(content)) {
		copy(snd.buffer[snd.getIdx(snd.LBW):], content)
	} else {
		// 2.(2) Otherwise, write twice
		// <1> write remainR bytes of content into right part of buffer
		copy(snd.buffer[snd.getIdx(snd.LBW):], content[:remainR])
		content2 := content[remainR:]
		// <2> write remainL part of content into left of buffer
		copy(snd.buffer, content2)
	}
	// 3. update total and LBW
	snd.LBW += bnum
	snd.total += bnum
	return bnum
}

// *********************************************************************************************
// LBW => Write bytes into send buffer
func (snd *SND) GetZeroProbe() ([]byte, uint32) {
	seqNum := snd.NXT
	// 1. length == 1
	len := uint32(1)
	// 2. copy payload
	payload := make([]byte, len)
	copy(payload, snd.buffer[snd.getIdx(snd.NXT):snd.getIdx(snd.NXT)+len])
	// 3. update metadata
	snd.NXT += len
	snd.WND -= len
	return payload, seqNum
}

func (snd *SND) GetSegment() ([]byte, uint32) {
	seqNum := snd.NXT

	mtu := uint32(proto.DEFAULT_PACKET_MTU - proto.DEFAULT_IPHDR_LEN - proto.DEFAULT_TCPHDR_LEN)
	len := getMinLength(mtu, snd.WND, (snd.LBW-1)-snd.NXT+1)
	// 1. Length of segment = min(mtu, snd.LBW - snd.NXT, snd.WIN)

	payload := make([]byte, len)
	if snd.getIdx(snd.NXT)+len < proto.BUFFER_SIZE {
		// 2. (1) If enough bytes on the right of SND, send them all
		// copy all, notice that there must be <, <= will cause range problem like [9:0]
		copy(payload, snd.buffer[snd.getIdx(snd.NXT):snd.getIdx(snd.NXT+len)])
	} else {
		// 2. (2) Otherwise, send bytes on the right and left
		// copy right and left
		len1 := proto.BUFFER_SIZE - snd.getIdx(snd.NXT)
		copy(payload, snd.buffer[snd.getIdx(snd.NXT):proto.BUFFER_SIZE])
		len2 := len - len1
		copy(payload[len1:], snd.buffer[:len2])
	}
	// 3. Update metadata of send buffer
	if snd.WND != 0 {
		snd.NXT += len
		snd.WND -= len
	}
	return payload, seqNum
}

// *********************************************************************************************
// Helper function
func (snd *SND) CanSend() bool {
	return snd.NXT < snd.LBW
}

func (snd *SND) UpdateWin(tcpHeaderWin uint16) {
	snd.WND = uint32(tcpHeaderWin)
}

// Check if current send buffer is full
func (snd *SND) IsFull() bool {
	return snd.total == proto.SND_BUFFER_SIZE
}

func (snd *SND) getRemainingBytes() uint32 {
	return proto.SND_BUFFER_SIZE - snd.total
}

// Get number of remaining bytes from LBW to End
func (snd *SND) getRemainingBytesRight() uint32 {
	return proto.SND_BUFFER_SIZE - 1 - snd.getIdx(snd.LBW) + 1
}

func (snd *SND) getIdx(seqNum uint32) uint32 {
	return (seqNum - snd.ISS - 1) % proto.SND_BUFFER_SIZE
}

func getMinLength(mtu, win, remainBytes uint32) uint32 {
	var min1 uint32
	var min2 uint32

	if mtu < win {
		min1 = mtu
	} else {
		min1 = win
	}

	if min1 < remainBytes {
		min2 = min1
	} else {
		min2 = remainBytes
	}
	return min2
}
