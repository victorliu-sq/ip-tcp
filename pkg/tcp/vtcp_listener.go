package tcp

import (
	"fmt"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"
)

type VTCPListener struct {
	ID        uint16
	state     string
	localPort uint16
	// listern.acceptLoop(), conn => spawnCh => VAccpet()
	ConnQueue chan *VTCPConn
	// listen.
	SegRcvChan  chan *proto.Segment
	ConnInQueue int
	CancelChan  chan bool
	CLIChan     chan *proto.NodeCLI
}

func NewListener(port uint16) *VTCPListener {
	listener := &VTCPListener{
		localPort:   port,
		state:       proto.LISTENER,
		ConnQueue:   make(chan *VTCPConn),
		SegRcvChan:  make(chan *proto.Segment),
		ConnInQueue: 0,
		CancelChan:  make(chan bool),
		CLIChan:     make(chan *proto.NodeCLI),
	}
	go listener.VListenerAcceptLoop()
	return listener
}

func (listener *VTCPListener) VListenerAcceptLoop() error {
	for {
		select {
		case <-listener.CancelChan:
			return nil
		case segment := <-listener.SegRcvChan:
			myDebug.Debugln("socket listening on %v receives a request from %v:%v",
				listener.localPort, segment.IPhdr.Src.String(), segment.TCPhdr.SrcPort)
			// Notice we need to reverse dst and stc in segment to create a new conn
			conn := NewNormalSocket(segment.TCPhdr.SeqNum, segment.TCPhdr.SrcPort, segment.TCPhdr.DstPort, segment.IPhdr.Src, segment.IPhdr.Dst)
			listener.ConnQueue <- conn
			listener.ConnInQueue++
		}
	}
}

func (listener *VTCPListener) VAccept() (*VTCPConn, error) {
	conn := <-listener.ConnQueue
	if conn == nil {
		return nil, fmt.Errorf("fail to produce a new socket")
	}
	listener.ConnInQueue--
	return conn, nil
}
