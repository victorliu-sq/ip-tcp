package transport

import (
	"fmt"
	"net"
	"tcpip/pkg/proto"
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
		ci := &ClientInfo{
			ISS:        segment.TCPhdr.SeqNum,
			RemotePort: segment.TCPhdr.SrcPort,
			LocalPort:  segment.TCPhdr.DstPort,
			RemoteAddr: segment.IPhdr.Src,
			LocalAddr:  segment.IPhdr.Dst,
		}
		go listener.SendToClientInfoChan(ci)
		fmt.Println("listener pushes one client info")
	}
}

func (listener *VTCPListener) SendToClientInfoChan(ci *ClientInfo) {
	listener.ClientInfoChan <- ci
}

func (listener *VTCPListener) VAccept() *VTCPConn {
	ci := <-listener.ClientInfoChan
	conn := listener.ST.CreateConnSYNRCV(ci.RemoteAddr.String(), ci.LocalAddr.String(), ci.RemotePort, ci.LocalPort, listener.NodeSegSendChan)
	return conn
}

func (listener *VTCPListener) VAcceptLoop() {
	for listener.State == proto.LISTENER {
		listener.VAccept()
	}
}
