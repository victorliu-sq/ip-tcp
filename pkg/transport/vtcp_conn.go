package transport

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"tcpip/pkg/proto"
	"time"

	"github.com/google/netstack/tcpip/header"
)

type VTCPConn struct {
	// mu    sync.Mutex
	State string
	ID    uint16
	ISN   uint32
	// seqNum     uint32
	// ackNum     uint32
	LocalAddr  net.IP
	LocalPort  uint16
	RemoteAddr net.IP
	RemotePort uint16
	// windowSize      uint16
	ConnSegRcvChan  chan *proto.Segment
	NodeSegSendChan chan *proto.Segment
	// Write Condition Variable
	// wcv sync.Cond
	// Send Buffer
	// scv sync.Cond
	// sb  *SendBuffer // send buffer
	snd *SND
	// Retransmission
	rtmQueue      chan *proto.Segment  // retransmission queue
	seq2timestamp map[uint32]time.Time // seq # of 1 segment to expiration time
	//Recv
	// NonEmptyCond *sync.Cond
	// RcvBuf       *RecvBuffer
	rcv *RCV
	// ZeroProbe
	zeroProbe bool
	recvFIN   bool
	Fd        *os.File
}

func NewVTCPConnSYNSENT(dstPort, srcPort uint16, dstIP, srcIP net.IP, id uint16, seqNumber uint32, state string, nodeSegSendChan chan *proto.Segment) *VTCPConn {
	conn := &VTCPConn{
		// mu:         sync.Mutex{},
		LocalPort:  srcPort,
		LocalAddr:  srcIP,
		RemoteAddr: dstIP,
		RemotePort: dstPort,
		ID:         id,
		State:      state,
		ISN:        GenerateRandomNumber(),
		// ackNum:     seqNumber + 1, // first ackNum can be 1 by giving seqNumber 0 (client --> NConn)
		// windowSize: DEFAULTWINDOWSIZE,
		NodeSegSendChan: nodeSegSendChan,
		ConnSegRcvChan:  make(chan *proto.Segment),
		// Retransmission
		rtmQueue:      make(chan *proto.Segment),
		seq2timestamp: make(map[uint32]time.Time),
		zeroProbe:     false,
		recvFIN:       false,
		Fd:            nil,
	}
	// conn.NonEmptyCond = sync.NewCond(&conn.mu)
	// Create SND
	conn.snd = NewSND(conn.ISN)
	go conn.ConnStateMachineLoop()
	// go conn.retransmissionLoop()
	return conn
}

func NewVTCPConnSYNRCV(dstPort, srcPort uint16, dstIP, srcIP net.IP, id uint16, seqNumber uint32, state string, nodeSegSendChan chan *proto.Segment) *VTCPConn {
	conn := &VTCPConn{
		// mu:         sync.Mutex{},
		LocalPort:  srcPort,
		LocalAddr:  srcIP,
		RemoteAddr: dstIP,
		RemotePort: dstPort,
		ID:         id,
		State:      state,
		ISN:        GenerateRandomNumber() - 1123456789,
		// ackNum:     seqNumber + 1, // first ackNum can be 1 by giving seqNumber 0 (client --> NConn)
		// windowSize: DEFAULTWINDOWSIZE,
		NodeSegSendChan: nodeSegSendChan,
		ConnSegRcvChan:  make(chan *proto.Segment),
		// Retransmission
		rtmQueue:      make(chan *proto.Segment),
		seq2timestamp: make(map[uint32]time.Time),
		zeroProbe:     false,
		recvFIN:       false,
		Fd:            nil,
	}
	// conn.NonEmptyCond = sync.NewCond(&conn.mu)
	// Create SND
	conn.snd = NewSND(conn.ISN)
	// Create RCV
	conn.rcv = NewRCV(seqNumber)
	go conn.ConnStateMachineLoop()
	// go conn.retransmissionLoop()
	return conn
}

func (conn *VTCPConn) FormTuple() string {
	remotePort := strconv.Itoa(int(conn.RemotePort))
	localPort := strconv.Itoa(int(conn.LocalPort))
	remoteAddr := conn.RemoteAddr.String()
	localAddr := conn.LocalAddr.String()
	return fmt.Sprintf("%v:%v:%v:%v", remoteAddr, remotePort, localAddr, localPort)
}

func (conn *VTCPConn) BuildTCPHdr(flags int, seqNum, ackNum uint32) *header.TCPFields {
	return &header.TCPFields{
		SrcPort:    conn.LocalPort,
		DstPort:    conn.RemotePort,
		SeqNum:     seqNum,
		AckNum:     ackNum,
		DataOffset: proto.DEFAULT_DATAOFFSET,
		Flags:      uint8(flags),
		// WindowSize:    conn.windowSize,
		Checksum:      0,
		UrgentPointer: 0,
	}
}
