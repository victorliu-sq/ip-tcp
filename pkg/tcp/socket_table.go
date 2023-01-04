package tcp

import (
	"fmt"
	"sync"
)

type SocketTable struct {
	mu sync.Mutex
	//tuple := remoteIP::remotePort::localPort
	ConnPort          uint16 // which port to assign for a new normal conn
	counter           uint16 // which id to assign for a new listen conn / normal conn
	id2Conns          map[uint16]*VTCPConn
	tuple2NormalConns map[string]*VTCPConn
	id2Listeners      map[uint16]*VTCPListener
	port2Listeners    map[uint16]*VTCPListener
}

func NewSocketTable() *SocketTable {
	return &SocketTable{
		ConnPort:          uint16(1024),
		counter:           uint16(0),
		id2Conns:          make(map[uint16]*VTCPConn),
		tuple2NormalConns: make(map[string]*VTCPConn),
		id2Listeners:      make(map[uint16]*VTCPListener),
		port2Listeners:    make(map[uint16]*VTCPListener),
	}
}

func (table *SocketTable) PrintSockets() {
	fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v\n", "socket", "local-addr", "port", "dst-addr", "port", "status")
	fmt.Println("--------------------------------------------------------------")
	// Print out Listener Conns
	for _, ls := range table.id2Listeners {
		fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v\n", ls.ID, "0.0.0.0", ls.localPort, "0.0.0.0", "0", ls.state)
	}
	for _, conn := range table.id2Conns {
		// 0       10.0.0.1        1024            10.0.0.14       80      ESTAB
		fmt.Printf("%-8v %-16v %-12v %-12v %-12v %-12v \n", conn.ID, conn.LocalAddr, conn.LocalPort, conn.RemoteAddr, conn.RemotePort, conn.state)
	}
}

func (table *SocketTable) OfferListener(port uint16) *VTCPListener {
	table.mu.Lock()
	defer table.mu.Unlock()
	listener := NewListener(port)
	listener.ID = table.counter
	table.port2Listeners[port] = listener
	table.id2Listeners[listener.ID] = listener
	// go listener.acceptLoop()
	table.counter++
	return listener
}

func (table *SocketTable) OfferConn(conn *VTCPConn) {
	table.mu.Lock()
	defer table.mu.Unlock()
	tuple := conn.GetTuple()
	conn.ID = table.counter
	table.id2Conns[conn.ID] = conn
	table.tuple2NormalConns[tuple] = conn
	table.counter++
	table.ConnPort++
}

func (table *SocketTable) DeleteSocket(id uint16) {
	if conn, found := table.id2Conns[id]; found {
		tuple := conn.GetTuple()
		delete(table.id2Conns, id)
		delete(table.tuple2NormalConns, tuple)
	}
	if ls, found := table.id2Listeners[id]; found {
		port := ls.localPort
		delete(table.id2Listeners, id)
		delete(table.port2Listeners, port)
	}
}

func (table *SocketTable) FindListener(port uint16) *VTCPListener {
	table.mu.Lock()
	defer table.mu.Unlock()
	return table.port2Listeners[port]
}

func (table *SocketTable) FindConn(tuple string) *VTCPConn {
	table.mu.Lock()
	defer table.mu.Unlock()
	return table.tuple2NormalConns[tuple]
}

func (table *SocketTable) FindConnByID(id uint16) *VTCPConn {
	table.mu.Lock()
	defer table.mu.Unlock()
	return table.id2Conns[id]
}
