package transport

import "tcpip/pkg/proto"

type RCV struct {
	buffer []byte
	ISS    uint32
	NXT    uint32
	LBR    uint32
	WND    uint32
}

func NewRCV(ClientISS uint32) *RCV {
	return &RCV{
		buffer: make([]byte, proto.RCV_BUFFER_SIZE),
		ISS:    ClientISS,
		NXT:    ClientISS + 1,
		LBR:    ClientISS + 1,
		WND:    proto.RCV_BUFFER_SIZE,
	}
}

func (rcv *RCV) PrintRCV() {
	DPrintf("----------------RCV----------------\n")
	DPrintf("%-16v %-16v %-16v\n", "ISS", "NXT", "LBR")
	DPrintf("%-16v %-16v %-16v\n", rcv.ISS, rcv.NXT, rcv.LBR)
}
