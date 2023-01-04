package tcp

import (
	"tcpip/pkg/proto"
	"time"

	"github.com/google/netstack/tcpip/header"
)

func (conn *VTCPConn) doFINWAIT1() {
	timeout := time.After(proto.RetranInterval)
	for {
		select {
		case <-timeout:
			conn.mu.Lock()
			conn.send2(FIN, "FIN_WAIT1 RESEND FIN")
			conn.mu.Unlock()
			timeout = time.After(proto.RetranInterval)
		case segRev := <-conn.SegRcvChan:
			flag := segRev.TCPhdr.Flags
			if flag == ACK {
				if segRev.TCPhdr.AckNum <= conn.seqNum {
					conn.HandleRcvSegInSendBuffer(segRev)
				}
				if segRev.TCPhdr.AckNum == conn.seqNum+1 {
					conn.mu.Lock()
					conn.seqNum++
					conn.state = proto.FINWAIT2
					conn.PrintIncoming(segRev, "Recv ACK -> FIN_WAIT2")
					go conn.doFINWAIT2()
					conn.mu.Unlock()
					return
				}
			}
			if flag == FIN {
				ackMyFIN := segRev.TCPhdr.AckNum == conn.seqNum+1
				recvAll := conn.ackNum == segRev.TCPhdr.SeqNum
				if !recvAll {
					conn.PrintIncoming(segRev, "Recv SIMUL FIN & MISSING DATA")
					conn.mu.Lock()
					conn.send2(ACK, "ACK SIMUL FIN BUT MISSING DATA")
					conn.mu.Unlock()
				} else {
					if ackMyFIN {
						conn.PrintIncoming(segRev, "Recv SIMUL FIN & ACK MY FIN")
						conn.mu.Lock()
						conn.ackNum++
						conn.state = proto.TIMEWAIT
						conn.send2(ACK, "ACK SIMUL FIN+ACK -> TIME WAIT")
						go conn.doTimeWait()
						conn.mu.Unlock()
					} else {
						conn.PrintIncoming(segRev, "Recv SIMUL FIN & NOT ACK MY FIN")
						conn.mu.Lock()
						conn.seqNum++
						conn.ackNum++
						conn.state = proto.CLOSING
						conn.send2(ACK, "ACK SIMUL FIN -> CLOSING")
						go conn.doClosing()
						conn.mu.Unlock()
					}
					return
				}
			}
		}
	}
}

func (conn *VTCPConn) doFINWAIT2() {
	for {
		segRev := <-conn.SegRcvChan
		finpkt := segRev.TCPhdr.Flags == FIN
		latestFin := conn.ackNum == segRev.TCPhdr.SeqNum
		if segRev.TCPhdr.Flags == header.TCPFlagAck {
			conn.handleACKSeg(segRev)
		}
		if finpkt && latestFin {
			conn.PrintIncoming(segRev, "RECV FIN")
			conn.mu.Lock()
			conn.ackNum++
			conn.send2(ACK, "ACK FIN -> TIME_W")
			conn.state = proto.TIMEWAIT
			go conn.doTimeWait()
			conn.mu.Unlock()
			return
		}
		conn.PrintIncoming(segRev, "maybe error")
	}
}

func (conn *VTCPConn) doLastAck() {
	timeout := time.After(proto.RetranInterval)
	for {
		select {
		case <-timeout:
			conn.mu.Lock()
			conn.send2(FIN, "CLOSE_W RESEND FIN")
			conn.mu.Unlock()
			timeout = time.After(proto.RetranInterval)
		case segRev := <-conn.SegRcvChan:
			conn.PrintIncoming(segRev, "RECV LAST ACK")
			isAck := segRev.TCPhdr.Flags == ACK
			ackMe := segRev.TCPhdr.AckNum-1 == conn.seqNum
			if isAck && ackMe {
				conn.CLIChan <- &proto.NodeCLI{CLIType: proto.CLI_DELETECONN, Val16: conn.ID}
				return
			}
		}
	}
}

func (conn *VTCPConn) doClosing() {
	timeout := time.After(proto.RetranInterval)
	for {
		select {
		case <-timeout:
			conn.mu.Lock()
			conn.send2(ACK, "CLOSING RESEND ACK")
			conn.mu.Unlock()
			timeout = time.After(proto.RetranInterval)
		case segRev := <-conn.SegRcvChan:
			conn.PrintIncoming(segRev, "RECV LAST ACK")
			isAck := segRev.TCPhdr.Flags == ACK
			isFin := segRev.TCPhdr.Flags == FIN
			ackMe := segRev.TCPhdr.AckNum-1 == conn.seqNum
			if isAck && ackMe {
				go conn.doTimeWait()
				return
			}
			if isFin {
				conn.mu.Lock()
				conn.send2(ACK, "CLOSING RESEND ACK")
				conn.mu.Unlock()
				timeout = time.After(proto.RetranInterval)
			}
		}
	}
}

func (conn *VTCPConn) doTimeWait() {
	//conn.BlockChan <- &proto.NodeCLI{CLIType: proto.CLI_CLOSESOCKET, Val16: conn.ID}
	timeout := time.After(proto.RetranInterval)
	for {
		select {
		case <-timeout:
			conn.CLIChan <- &proto.NodeCLI{CLIType: proto.CLI_DELETECONN, Val16: conn.ID}
			return
		case <-conn.SegRcvChan:
			conn.mu.Lock()
			conn.send2(ACK, "TIME WAIT RESEND ACK")
			conn.mu.Unlock()
			go conn.doTimeWait()
			return
		}
	}
}
