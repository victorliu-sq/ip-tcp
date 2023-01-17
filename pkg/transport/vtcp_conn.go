package transport

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"tcpip/pkg/proto"
	"time"

	"github.com/google/netstack/tcpip/header"
)

type VTCPConn struct {
	mu              sync.Mutex
	State           string
	ID              uint16
	ISN             uint32
	LocalAddr       net.IP
	LocalPort       uint16
	RemoteAddr      net.IP
	RemotePort      uint16
	ConnSegRcvChan  chan *proto.Segment
	NodeSegSendChan chan *proto.Segment
	// Write Condition Variable
	wcond sync.Cond
	// Send Buffer
	scond sync.Cond
	snd   *SND
	// Retransmission
	retsmChan     chan *proto.Segment  // retransmission queue
	seq2timestamp map[uint32]time.Time // seq # of 1 segment to expiration time
	// RCV Buffer
	rcv   *RCV
	rcond sync.Cond
	// ZeroProbe
	zeroProbe bool
	// 2MSLTimeout
	timeout2MSL time.Time
}

func NewVTCPConnSYNSENT(dstPort, srcPort uint16, dstIP, srcIP net.IP, id uint16, seqNumber uint32, state string, nodeSegSendChan chan *proto.Segment) *VTCPConn {
	conn := &VTCPConn{
		mu:              sync.Mutex{},
		LocalPort:       srcPort,
		LocalAddr:       srcIP,
		RemoteAddr:      dstIP,
		RemotePort:      dstPort,
		ID:              id,
		State:           state,
		ISN:             GenerateRandomNumber(),
		NodeSegSendChan: nodeSegSendChan,
		ConnSegRcvChan:  make(chan *proto.Segment),
		// Retransmission
		retsmChan:     make(chan *proto.Segment),
		seq2timestamp: make(map[uint32]time.Time),
		zeroProbe:     false,
	}
	conn.wcond = *sync.NewCond(&conn.mu)
	conn.scond = *sync.NewCond(&conn.mu)
	conn.rcond = *sync.NewCond(&conn.mu)
	// Create SND
	conn.snd = NewSND(conn.ISN)
	go conn.ConnStateMachineLoop()
	go conn.RetransmissionLoop()
	return conn
}

func NewVTCPConnSYNRCV(dstPort, srcPort uint16, dstIP, srcIP net.IP, id uint16, seqNumber uint32, state string, nodeSegSendChan chan *proto.Segment) *VTCPConn {
	conn := &VTCPConn{
		// mu:         sync.Mutex{},
		LocalPort:       srcPort,
		LocalAddr:       srcIP,
		RemoteAddr:      dstIP,
		RemotePort:      dstPort,
		ID:              id,
		State:           state,
		ISN:             GenerateRandomNumber() - 1123456789,
		NodeSegSendChan: nodeSegSendChan,
		ConnSegRcvChan:  make(chan *proto.Segment),
		// Retransmission
		retsmChan:     make(chan *proto.Segment),
		seq2timestamp: make(map[uint32]time.Time),
		zeroProbe:     false,
	}
	conn.wcond = *sync.NewCond(&conn.mu)
	conn.scond = *sync.NewCond(&conn.mu)
	conn.rcond = *sync.NewCond(&conn.mu)
	// Create SND
	conn.snd = NewSND(conn.ISN)
	// Create RCV
	conn.rcv = NewRCV(seqNumber)
	go conn.ConnStateMachineLoop()
	go conn.RetransmissionLoop()
	return conn
}

// *******************************************************************
// Normal socket State machine
func (conn *VTCPConn) ConnStateMachineLoop() {
	for {
		segment := <-conn.ConnSegRcvChan
		// fmt.Println(segment)
		// DPrintf("Conn [%v] Receive one segment with seqNum %v \n", conn.State, segment.TCPhdr.SeqNum)
		switch conn.State {
		case proto.SYN_SENT:
			conn.HandleSegmentInStateSYNSENT(segment)
		case proto.SYN_RCVD:
			conn.HandleSegmentInStateSYNRCVD(segment)
		case proto.ESTABLISH:
			conn.HandleSegmentInStateESTABLISH(segment)
		case proto.FINWAIT1:
			conn.HandleSegmentInStateFINWAIT1(segment)
		case proto.FINWAIT2:
			conn.HandleSegmentInStateFINWAIT2(segment)
		case proto.TIMEWAIT:
			conn.HandleSegmentInStateCLOSEWAIT(segment)
		case proto.CLOSEWAIT:
			conn.HandleSegmentInStateCLOSEWAIT(segment)
		case proto.LASTACK:
			conn.HandleSegmentInStateLASTACK(segment)
		}
	}
}

func (conn *VTCPConn) FormTuple() string {
	remotePort := strconv.Itoa(int(conn.RemotePort))
	localPort := strconv.Itoa(int(conn.LocalPort))
	remoteAddr := conn.RemoteAddr.String()
	localAddr := conn.LocalAddr.String()
	return fmt.Sprintf("%v:%v:%v:%v", remoteAddr, remotePort, localAddr, localPort)
}

func (conn *VTCPConn) BuildTCPHdr(flags int, seqNum, ackNum, win uint32) *header.TCPFields {
	return &header.TCPFields{
		SrcPort:       conn.LocalPort,
		DstPort:       conn.RemotePort,
		SeqNum:        seqNum,
		AckNum:        ackNum,
		DataOffset:    proto.DEFAULT_DATAOFFSET,
		Flags:         uint8(flags),
		WindowSize:    uint16(win),
		Checksum:      0,
		UrgentPointer: 0,
	}
}

func (conn *VTCPConn) GetState() string {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	return conn.State
}
