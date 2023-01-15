package transport

import (
	"fmt"
	"net"
	"sync"
	"tcpip/pkg/proto"
)

type SocketTable struct {
	// socketTable needs mutex because multiple listeners can access it to create new normal sockets
	mu sync.Mutex
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
		mu:                  sync.Mutex{},
		connPort:            uint16(1024),
		counter:             uint16(0),
		id2Conns:            make(map[uint16]*VTCPConn),
		tuple2Conns:         make(map[string]*VTCPConn),
		id2Listeners:        make(map[uint16]*VTCPListener),
		localPort2Listeners: make(map[uint16]*VTCPListener),
	}
}

func (st *SocketTable) PrintSockets() {
	st.mu.Lock()
	defer st.mu.Unlock()
	fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v\n", "socket", "local-addr", "port", "dst-addr", "port", "status")
	fmt.Println("--------------------------------------------------------------")
	// Print out Listener Conns
	for _, listener := range st.id2Listeners {
		fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v\n", listener.ID, "0.0.0.0", listener.LocalPort, "0.0.0.0", "0", listener.State)
	}
	for _, conn := range st.id2Conns {
		// 0       10.0.0.1        1024            10.0.0.14       80      EstAB
		fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v \n", conn.ID, conn.LocalAddr, conn.LocalPort, conn.RemoteAddr, conn.RemotePort, conn.State)
	}
}

// ***********************************************************************************************
// VTCPListener
func (st *SocketTable) CreateListener(port uint16, nodeSegSendChan chan *proto.Segment) *VTCPListener {
	st.mu.Lock()
	defer st.mu.Unlock()
	listener := NewVTCPListener(port, st.counter, st, nodeSegSendChan)
	st.id2Listeners[listener.ID] = listener
	st.localPort2Listeners[listener.LocalPort] = listener
	st.counter++
	return listener
}

func (st *SocketTable) Port2Listener(port uint16) (*VTCPListener, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	listener, ok := st.localPort2Listeners[port]
	return listener, ok
}

// ***********************************************************************************************
// VTCPConn
func (st *SocketTable) CreateConnSYNSENT(remoteAddr, localAddr string, remotePort uint16, nodeSegSendChan chan *proto.Segment) *VTCPConn {
	st.mu.Lock()
	defer st.mu.Unlock()
	conn := NewVTCPConnSYNSENT(remotePort, st.connPort, net.ParseIP(remoteAddr), net.ParseIP(localAddr), st.counter, uint32(st.counter), proto.SYN_SENT, nodeSegSendChan)
	// Update Socket Table
	tuple := conn.FormTuple()
	// fmt.Println(tuple)
	st.id2Conns[conn.ID] = conn
	st.tuple2Conns[tuple] = conn
	st.counter++
	st.connPort++
	return conn
}

func (st *SocketTable) CreateConnSYNRCV(remoteAddr, localAddr string, remotePort, localPort uint16, seqNum uint32, nodeSegSendChan chan *proto.Segment) *VTCPConn {
	st.mu.Lock()
	defer st.mu.Unlock()
	conn := NewVTCPConnSYNRCV(remotePort, localPort, net.ParseIP(remoteAddr), net.ParseIP(localAddr), st.counter, seqNum, proto.SYN_RCVD, nodeSegSendChan)
	// Update Socket Table
	tuple := conn.FormTuple()
	// fmt.Println(tuple)
	st.id2Conns[conn.ID] = conn
	st.tuple2Conns[tuple] = conn
	st.counter++
	st.connPort++
	return conn
}

func (st *SocketTable) Tuple2Conn(tuple string) (*VTCPConn, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	conn, ok := st.tuple2Conns[tuple]
	return conn, ok
}

func (st *SocketTable) ID2Conn(id uint16) (*VTCPConn, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	conn, ok := st.id2Conns[id]
	return conn, ok
}
