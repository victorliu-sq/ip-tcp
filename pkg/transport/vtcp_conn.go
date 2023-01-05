package transport

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"tcpip/pkg/proto"
	"time"

	"github.com/google/netstack/tcpip/header"
)

type VTCPConn struct {
	mu         sync.Mutex
	State      string
	ID         uint16
	ISS        uint32
	seqNum     uint32
	ackNum     uint32
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
	// Retransmission
	rtmQueue      chan *proto.Segment  // retransmission queue
	seq2timestamp map[uint32]time.Time // seq # of 1 segment to expiration time
	//Recv
	NonEmptyCond *sync.Cond
	// RcvBuf       *RecvBuffer
	// ZeroProbe
	zeroProbe bool
	recvFIN   bool
	Fd        *os.File
}

func NewVTCPConn(dstPort, srcPort uint16, dstIP, srcIP net.IP, id uint16, seqNumber uint32, state string, nodeSegSendChan chan *proto.Segment) *VTCPConn {
	conn := &VTCPConn{
		mu:         sync.Mutex{},
		LocalPort:  srcPort,
		LocalAddr:  srcIP,
		RemoteAddr: dstIP,
		RemotePort: dstPort,
		ID:         id,
		State:      state,
		ISS:        rand.Uint32(),
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
	conn.NonEmptyCond = sync.NewCond(&conn.mu)

	go conn.ConnSegmentHandler()
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

func (conn *VTCPConn) ConnSegmentHandler() {
	for {
		segment := <-conn.ConnSegRcvChan
		fmt.Println(segment)
	}
}

func (conn *VTCPConn) BuildTCPHdr(flags int, seqNum uint32) *header.TCPFields {
	return &header.TCPFields{
		SrcPort:    conn.LocalPort,
		DstPort:    conn.RemotePort,
		SeqNum:     seqNum,
		AckNum:     conn.ackNum,
		DataOffset: proto.DEFAULT_DATAOFFSET,
		Flags:      uint8(flags),
		// WindowSize:    conn.windowSize,
		Checksum:      0,
		UrgentPointer: 0,
	}
}
