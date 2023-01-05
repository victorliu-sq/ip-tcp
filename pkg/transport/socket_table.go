package transport

import (
	"fmt"
	"net"
	"tcpip/pkg/proto"
)

type SocketTable struct {
	// which port to assign for a new normal conn
	connPort uint16
	// which id to assign for a new listen conn / normal conn
	counter uint16
	// normal socket => mapped by tuple
	id2Conns    map[uint16]*VTCPConn
	tuple2Conns map[string]*VTCPConn
	// listen socket => mapped by local port
	id2Listeners        map[uint16]*VTCPListener
	localPort2Listeners map[uint16]*VTCPListener
}

func NewSocketTable() *SocketTable {
	return &SocketTable{
		connPort:            uint16(1024),
		counter:             uint16(0),
		id2Conns:            make(map[uint16]*VTCPConn),
		tuple2Conns:         make(map[string]*VTCPConn),
		id2Listeners:        make(map[uint16]*VTCPListener),
		localPort2Listeners: make(map[uint16]*VTCPListener),
	}
}

func (table *SocketTable) PrintSockets() {
	fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v\n", "socket", "local-addr", "port", "dst-addr", "port", "status")
	fmt.Println("--------------------------------------------------------------")
	// Print out Listener Conns
	for _, listener := range table.id2Listeners {
		fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v\n", listener.ID, "0.0.0.0", listener.LocalPort, "0.0.0.0", "0", listener.State)
	}
	for _, conn := range table.id2Conns {
		// 0       10.0.0.1        1024            10.0.0.14       80      ESTAB
		fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v \n", conn.ID, conn.LocalAddr, conn.LocalPort, conn.RemoteAddr, conn.RemotePort, conn.State)
	}
}

// ***********************************************************************************************
// VTCPListener
func (ST *SocketTable) CreateListener(port uint16) *VTCPListener {
	listener := NewVTCPListener(port, ST.counter)
	ST.id2Listeners[listener.ID] = listener
	ST.localPort2Listeners[listener.LocalPort] = listener
	ST.counter++
	return listener
}

func (ST *SocketTable) Port2Listener(port uint16) (*VTCPListener, bool) {
	listener, ok := ST.localPort2Listeners[port]
	return listener, ok
}

// ***********************************************************************************************
// VTCPConn
func (ST *SocketTable) CreateConn(remoteAddr, localAddr string, remotePort uint16, nodeSegSendChan chan *proto.Segment) *VTCPConn {
	conn := NewVTCPConn(remotePort, ST.connPort, net.ParseIP(remoteAddr), net.ParseIP(localAddr), 0, uint32(ST.counter), proto.SYN_SENT, nodeSegSendChan)
	tuple := conn.FormTuple()
	// fmt.Println(tuple)
	ST.id2Conns[conn.ID] = conn
	ST.tuple2Conns[tuple] = conn
	// conn.NodeSegSendChan = node.segSendChan
	// conn.CLIChan = node.NodeCLIChan
	ST.counter++
	ST.connPort++
	return conn
}

func (ST *SocketTable) Tuple2Conn(tuple string) (*VTCPConn, bool) {
	conn, ok := ST.tuple2Conns[tuple]
	return conn, ok
}
