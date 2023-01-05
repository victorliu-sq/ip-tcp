package tcp

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sync"
	"tcpip/pkg/proto"
	"time"

	"github.com/google/netstack/tcpip/header"
)

const (
	FIN               = header.TCPFlagFin | header.TCPFlagAck
	ACK               = header.TCPFlagAck
	MINACKNUM         = 1
	DEFAULTDATAOFFSET = 20
	DEFAULTWINDOWSIZE = 65535
)

type VTCPConn struct {
	mu              sync.Mutex
	state           string
	seqNum          uint32
	ackNum          uint32
	LocalAddr       net.IP
	LocalPort       uint16
	RemoteAddr      net.IP
	RemotePort      uint16
	windowSize      uint16
	ID              uint16
	SegRcvChan      chan *proto.Segment
	NodeSegSendChan chan *proto.Segment
	// Write Condition Variable
	wcv sync.Cond
	// Send Buffer
	scv sync.Cond
	sb  *SendBuffer // send buffer
	// Retransmission
	rtmQueue      chan *proto.Segment  // retransmission queue
	seq2timestamp map[uint32]time.Time // seq # of 1 segment to expiration time
	//Recv
	NonEmptyCond *sync.Cond
	EstabCond    *sync.Cond
	RcvBuf       *RecvBuffer
	BlockChan    chan *proto.NodeCLI
	cancelChan   chan bool
	CLIChan      chan *proto.NodeCLI
	CloseChan    chan bool
	// ZeroProbe
	zeroProbe bool
	recvFIN   bool
	Fd        *os.File
}

func NewNormalSocket(seqNumber uint32, dstPort, srcPort uint16, dstIP, srcIP net.IP) *VTCPConn {
	conn := &VTCPConn{
		mu:         sync.Mutex{},
		state:      proto.SYN_RECV,
		seqNum:     rand.Uint32(),
		ackNum:     seqNumber + 1, // first ackNum can be 1 by giving seqNumber 0 (client --> NConn)
		LocalPort:  srcPort,
		LocalAddr:  srcIP,
		RemoteAddr: dstIP,
		RemotePort: dstPort,
		windowSize: DEFAULTWINDOWSIZE,
		SegRcvChan: make(chan *proto.Segment),
		// Retransmission
		rtmQueue:      make(chan *proto.Segment),
		seq2timestamp: make(map[uint32]time.Time),
		zeroProbe:     false,
		cancelChan:    make(chan bool),
		recvFIN:       false,
		CloseChan:     make(chan bool),
		Fd:            nil,
	}
	conn.NonEmptyCond = sync.NewCond(&conn.mu)
	conn.EstabCond = sync.NewCond(&conn.mu)
	go conn.retransmissionLoop()
	return conn
}

// ********************************************************************************************
// Client
func (conn *VTCPConn) SynSend() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// [HandShake1] Send Syn
	conn.sb = NewSendBuffer(conn.seqNum, DEFAULTWINDOWSIZE)
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	conn.rtmQueue <- seg
	// [Handshake2] Rev Syn+ACK
	for {
		conn.mu.Unlock()
		segRev := <-conn.SegRcvChan
		conn.mu.Lock()
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			// [Handshake3] Send Ack
			conn.seqNum = segRev.TCPhdr.AckNum
			conn.ackNum = segRev.TCPhdr.SeqNum + 1
			seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
			conn.NodeSegSendChan <- seg
			conn.state = proto.ESTABLISH
			// [Client] Create send buffer
			conn.sb.win = uint32(segRev.TCPhdr.WindowSize)
			conn.scv = *sync.NewCond(&conn.mu)
			conn.wcv = *sync.NewCond(&conn.mu)
			go conn.VSBufferSend()
			// go conn.VSBufferRcv()
			// [Double: Client] Create rcv buffer
			conn.RcvBuf = NewRecvBuffer(conn.ackNum, DEFAULTWINDOWSIZE)
			// [Client] Rev Segments
			go conn.estabRevAndSend()
			if conn.Fd != nil {
				conn.VSBufferWriteFile()
			}
			return
		}
	}
}

// ********************************************************************************************
// Server
func (conn *VTCPConn) SynRev() {
	// [Handshake2] Send Syn|ACK
	conn.mu.Lock()
	conn.seqNum -= 1000000000
	conn.seqNum++
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagSyn|header.TCPFlagAck, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	conn.rtmQueue <- seg
	conn.mu.Unlock()

	// [Handshake3] Rev ACK
	for {
		segRev := <-conn.SegRcvChan
		if conn.seqNum+1 == segRev.TCPhdr.AckNum {
			conn.seqNum = segRev.TCPhdr.AckNum
			conn.ackNum = segRev.TCPhdr.SeqNum
			conn.state = proto.ESTABLISH
			conn.RcvBuf.head = segRev.TCPhdr.SeqNum
			conn.RcvBuf.una = segRev.TCPhdr.SeqNum
			// [Server] Create rcv buffer
			// go conn.estabRev()
			// [Server] Rev Segments
			go conn.estabRevAndSend()
			// [Double: Server] Create send buffer
			conn.sb = NewSendBuffer(conn.seqNum, uint32(segRev.TCPhdr.WindowSize))
			conn.scv = *sync.NewCond(&conn.mu)
			conn.wcv = *sync.NewCond(&conn.mu)
			go conn.VSBufferSend()
			return
		}
	}
}

// ********************************************************************************************
// Handle SegRcv in both send buffer and rcv buffer
func (conn *VTCPConn) estabRevAndSend() {
	for {
		select {
		case segRev := <-conn.SegRcvChan:
			// it is possible ACK is lost and we get another SynAck
			if len(segRev.Payload) == 0 {
				// Rcv segments In Send Buffer
				conn.HandleRcvSegInSendBuffer(segRev)
			} else {
				// Rcv segments In Rcv Buffer
				conn.HandleRcvSegInRcvBuffer(segRev)
			}
		case <-conn.cancelChan:
			switch conn.state {
			case proto.ESTABLISH:
				conn.mu.Lock()
				defer conn.mu.Unlock()
				seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck|header.TCPFlagFin, conn.seqNum), []byte{})
				conn.NodeSegSendChan <- seg
				conn.state = proto.FINWAIT1
				go conn.doFINWAIT1()
				return
			default:
			}
		}
	}
}

func (conn *VTCPConn) handleACKSeg(seg *proto.Segment) {
	if len(seg.Payload) == 0 {
		conn.HandleRcvSegInSendBuffer(seg)
	} else {
		conn.HandleRcvSegInRcvBuffer(seg)
	}
}

// ********************************************************************************************
// Send TCP Packet through Established Normal Conn

func (conn *VTCPConn) VSBufferWrite(content []byte) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// fmt.Println("Hello")
	total := uint32(len(content))
	for total > 0 {
		// fmt.Printf("Hello1, isFull: %v\n", conn.sb.IsFull())
		if !conn.sb.IsFull() {
			bnum := conn.sb.WriteIntoBuffer(content)
			total -= bnum
			content = content[bnum:]
			conn.scv.Signal()
			// fmt.Println("Hello2")
		} else {
			conn.wcv.Wait()
		}
	}
}

func (conn *VTCPConn) VSBufferWriteFile() {

	fd := conn.Fd
	reader := bufio.NewReader(fd)
	content := make([]byte, conn.sb.win)
	num2Send, err := reader.Read(content)

	for err == nil {

		for num2Send > 0 {
			// fmt.Printf("Hello1, isFull: %v\n", conn.sb.IsFull())
			if !conn.sb.IsFull() {
				bnum := conn.sb.WriteIntoBuffer(content)
				num2Send -= int(bnum)
				content = content[bnum:]
				conn.scv.Signal()
				// fmt.Println("Hello2")
			} else {
				conn.wcv.Wait()
			}
		}

	}
	if err != io.EOF {
		fmt.Println(err)
	}
	conn.CloseChan <- true
	fd.Close()
}

func (conn *VTCPConn) VSBufferSend() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// mtu := int(uint32(proto.DEFAULTIPHDRLEN) + uint32(proto.DEFAULTTCPHDRLEN) + uint32(conn.sb.win))
	mtu := proto.DEFAULTPACKETMTU - proto.DEFAULTIPHDRLEN - proto.DEFAULTTCPHDRLEN
	for conn.state == proto.ESTABLISH {
		if conn.sb.CanSend() && !conn.zeroProbe {
			if conn.sb.win == 0 {
				payload, seqNum := conn.sb.GetZeroProbe()
				conn.send(payload, seqNum)
				conn.zeroProbe = true
			} else {
				// Get one segment, send it out and add it to retransmission queue
				payload, seqNum := conn.sb.GetSegmentToSendAndUpdateNxt(mtu)
				conn.send(payload, seqNum)
			}
		} else {
			conn.scv.Wait()
		}
	}
}

func (conn *VTCPConn) HandleRcvSegInSendBuffer(segRev *proto.Segment) {
	// It is possible we still get another Syn|Ack Segment
	if segRev.TCPhdr.Flags == (header.TCPFlagSyn | header.TCPFlagAck) {
		seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
		conn.NodeSegSendChan <- seg
		fmt.Printf("[HandShake3] Handshake Msg -> Send back another ACK %v\n", seg.TCPhdr.AckNum)
		return
	}

	conn.mu.Lock()
	acked := conn.sb.UpdateUNA(segRev)
	conn.wcv.Signal()
	conn.sb.UpdateWin(segRev.TCPhdr.WindowSize)
	if segRev.TCPhdr.WindowSize > 0 {
		conn.zeroProbe = false
		conn.scv.Signal()
	}
	conn.seqNum += acked
	// myDebug.Debugln("[SendBuffer_RevACK] %v send buffer remaing bytes %v", conn.LocalAddr.String(), conn.sb.GetRemainBytes())
	conn.mu.Unlock()
}

// ********************************************************************************************
// Retransmission Queue
func (conn *VTCPConn) retransmissionLoop() {
	for {
		segR := <-conn.rtmQueue
		if segR.TCPhdr.Flags == header.TCPFlagAck && len(segR.Payload) > 0 {
			go conn.retransmit(segR)
		} else if segR.TCPhdr.Flags == header.TCPFlagSyn || (segR.TCPhdr.Flags == (header.TCPFlagSyn | header.TCPFlagAck)) {
			go conn.retransmitHS(segR)
		}

	}
}

func (conn *VTCPConn) retransmitHS(segR *proto.Segment) {
	time.Sleep(proto.RetranInterval)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// for handshake segments, seq number should increment by 1
	if conn.seqNum >= segR.TCPhdr.SeqNum+1 {
		fmt.Printf("[Client] has successfully retransmitted 1 HS segment flag: %v\n", segR.TCPhdr.Flags)
		return
	}
	conn.NodeSegSendChan <- segR
	conn.rtmQueue <- segR
}

func (conn *VTCPConn) retransmit(segR *proto.Segment) {
	time.Sleep(300 * time.Millisecond)
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// for ACK segments, seq number should increment by len(payload)
	// if conn.seqNum < conn.sb.una {
	// 	return
	// }
	if conn.seqNum >= segR.TCPhdr.SeqNum+uint32(len(segR.Payload)) {
		return
	}
	// retransmit if not acked
	conn.NodeSegSendChan <- segR
	conn.rtmQueue <- segR
}

// ********************************************************************************************
// Recv
// func (conn *VTCPConn) estabRev() {
// 	for {
// 		segRev := <-conn.SegRcvChan
// 		conn.HandleRcvSegInRcvBuffer(segRev)
// 	}
// }

func (conn *VTCPConn) HandleRcvSegInRcvBuffer(segRev *proto.Segment) {
	conn.mu.Lock()
	if conn.windowSize == 0 {
		seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
		conn.NodeSegSendChan <- seg
		conn.mu.Unlock()
		return
	}
	status := conn.RcvBuf.GetSegStatus(segRev)
	// fmt.Println("status:", status)
	if status == OUTSIDEWINDOW {
		// bug_fix: unlock when call return
		conn.mu.Unlock()
		return
	}
	if status == UNDEFINED {
		conn.mu.Unlock()
		return
	}
	ackNum, windowSize := conn.RcvBuf.WriteSeg2Buf(segRev)
	headAcked := conn.RcvBuf.IsHeadAcked()
	if headAcked {
		conn.ackNum = ackNum
		conn.windowSize = windowSize
		conn.NonEmptyCond.Broadcast()
	}

	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
	conn.mu.Unlock()
}

func (conn *VTCPConn) Retriv(numBytes uint32, isBlock bool) {
	res := []byte{}
	totalRead := uint32(0)
	toRead := numBytes
	conn.CLIChan <- &proto.NodeCLI{CLIType: proto.CLI_BLOCKCLI}
	for {
		conn.mu.Lock()
		fmt.Println("reach 376")
		for !conn.RcvBuf.IsHeadAcked() {
			conn.NonEmptyCond.Wait()
		}
		if conn.recvFIN {
			break
		}
		output, numRead := conn.RcvBuf.ReadBuf(toRead)
		conn.windowSize += numRead
		conn.RcvBuf.SetWindowSize(uint32(conn.windowSize))
		res = append(res, output...)
		totalRead += uint32(numRead)
		// if string(res) == proto.TestString {
		// 	println("************************************************")
		// 	fmt.Println("Woww!!!!!!!!")
		// 	println("************************************************")
		// }
		conn.mu.Unlock()
		toRead -= uint32(numRead)
		if !isBlock || totalRead == numBytes {
			break
		}
	}
	conn.CLIChan <- &proto.NodeCLI{CLIType: proto.CLI_UNBLOCKCLI}
}

func (conn *VTCPConn) RetrivFile(fd *os.File) {
	res := []byte{}
	conn.CLIChan <- &proto.NodeCLI{CLIType: proto.CLI_BLOCKCLI}
	for {
		conn.mu.Lock()
		for !conn.RcvBuf.IsHeadAcked() {
			conn.NonEmptyCond.Wait()
		}
		if conn.recvFIN {
			conn.RcvBuf.una--
		}
		output, numRead := conn.RcvBuf.ReadBuf(DEFAULTWINDOWSIZE)
		conn.windowSize += numRead
		conn.RcvBuf.SetWindowSize(uint32(conn.windowSize))
		res = append(res, output...)
		conn.mu.Unlock()
		if conn.recvFIN {
			break
		}
	}
	conn.CLIChan <- &proto.NodeCLI{CLIType: proto.CLI_UNBLOCKCLI}
	fd.Write(res)
	fd.Close()
}

func (conn *VTCPConn) Close() {
	conn.cancelChan <- true
}

// ********************************************************************************************
// helper function
func (conn *VTCPConn) send(content []byte, seqNum uint32) {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(header.TCPFlagAck, seqNum), content)
	conn.NodeSegSendChan <- seg
	// add the segment to retransmission queue
	conn.rtmQueue <- seg
}

func (conn *VTCPConn) send2(flags int, info string) {
	seg := proto.NewSegment(conn.LocalAddr.String(), conn.RemoteAddr.String(), conn.buildTCPHdr(flags, conn.seqNum), []byte{})
	conn.NodeSegSendChan <- seg
}

func (conn *VTCPConn) GetTuple() string {
	return fmt.Sprintf("%v:%v:%v:%v", conn.RemoteAddr.String(), conn.RemotePort,
		conn.LocalAddr.String(), conn.LocalPort)
}

func (conn *VTCPConn) buildTCPHdr(flags int, seqNum uint32) *header.TCPFields {
	return &header.TCPFields{
		SrcPort:       conn.LocalPort,
		DstPort:       conn.RemotePort,
		SeqNum:        seqNum,
		AckNum:        conn.ackNum,
		DataOffset:    DEFAULTDATAOFFSET,
		Flags:         uint8(flags),
		WindowSize:    conn.windowSize,
		Checksum:      0,
		UrgentPointer: 0,
	}
}

func (conn *VTCPConn) GetAck() uint32 {
	return conn.ackNum
}
