package transport

import "tcpip/pkg/proto"

type SND struct {
	buffer []byte
	ISS    uint32
	UNA    uint32
	NXT    uint32
	LBW    uint32
	WIN    uint32
}

func NewSND(ISN uint32) *SND {
	snd := &SND{
		buffer: make([]byte, proto.SND_BUFFER_SIZE),
		ISS:    ISN,
		UNA:    ISN,
		NXT:    ISN + 1,
		LBW:    ISN + 1,
	}
	return snd
}

func (snd *SND) PrintSND() {
	DPrintf("----------------SND----------------\n")
	DPrintf("%-16v %-16v %-16v %-16v\n", "ISS", "UNA", "NXT", "LBW")
	DPrintf("%-16v %-16v %-16v %-16v\n", snd.ISS, snd.UNA, snd.NXT, snd.LBW)
}

func (snd *SND) AdvanceUNA(ackNum uint32) {
	snd.UNA = ackNum
}
