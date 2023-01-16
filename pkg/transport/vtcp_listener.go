package transport

import (
	"net"
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

type VTCPListener struct {
	State     string
	ID        uint16
	LocalPort uint16
	// listener has no remote address or remote port
	// any local address can map listener
	ListenerSegRcvChan chan *proto.Segment
	ClientInfoChan     chan *ClientInfo
	// we need these two metadata to create normal connection
	ST              *SocketTable
	NodeSegSendChan chan *proto.Segment
}

type ClientInfo struct {
	ISS        uint32
	LocalAddr  net.IP
	LocalPort  uint16
	RemoteAddr net.IP
	RemotePort uint16
}

func NewVTCPListener(port, id uint16, st *SocketTable, nodeSegSendChan chan *proto.Segment) *VTCPListener {
	listener := &VTCPListener{
		LocalPort:          port,
		ID:                 id,
		State:              proto.LISTENER,
		ClientInfoChan:     make(chan *ClientInfo),
		ListenerSegRcvChan: make(chan *proto.Segment),
		ST:                 st,
		NodeSegSendChan:    nodeSegSendChan,
	}
	go listener.ListenerSegHandler()
	return listener
}

func (listener *VTCPListener) ListenerSegHandler() {
	for {
		segment := <-listener.ListenerSegRcvChan
		// Check SYN Flag
		if segment.TCPhdr.Flags|header.TCPFlagSyn != 0 {
			ci := &ClientInfo{
				ISS:        segment.TCPhdr.SeqNum,
				RemotePort: segment.TCPhdr.SrcPort,
				LocalPort:  segment.TCPhdr.DstPort,
				RemoteAddr: segment.IPhdr.Src,
				LocalAddr:  segment.IPhdr.Dst,
			}
			go listener.SendToClientInfoChan(ci)
			DPrintf("listener pushes one client info\n")
		}
	}
}

func (listener *VTCPListener) SendToClientInfoChan(ci *ClientInfo) {
	listener.ClientInfoChan <- ci
}

func (listener *VTCPListener) VAccept() *VTCPConn {
	ci := <-listener.ClientInfoChan
	conn := listener.ST.CreateConnSYNRCV(ci.RemoteAddr.String(), ci.LocalAddr.String(), ci.RemotePort, ci.LocalPort, ci.ISS, listener.NodeSegSendChan)
	// Send a Segment with Flag (SYN | ACK)
	conn.SendSeg3WHS_SYNACK()
	return conn
}

func (listener *VTCPListener) VAcceptLoop() {
	// State Machine for Listen Socket
	for listener.State == proto.LISTENER {
		// Accept one Client Information and Create a new Normal Socket in State SYN_RCVD
		listener.VAccept()
		DPrintf("Listener [%v] has created one normal connection %v \n", listener.State, listener.LocalPort)
	}
}
