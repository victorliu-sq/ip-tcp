package network

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"tcpip/pkg/link"
	"tcpip/pkg/proto"
	"time"
)

type RoutingTable struct {
	// Interfaces
	ID2Interface map[uint8]*link.LinkInterface // Store interfaces and facilitate down and up
	// expiration channel
	NodeExChan chan *proto.NodeEx // Handle expiration of route
	// Routes
	DestIP2Route map[string]Route //store routes and facilitate finding target route for Test Packet
	// Node TCP Segment Channel
	NodeSegRevChan chan *proto.Segment
	// RIP metadata
	LocalIPSet         map[string]bool      // store all local IP in this rt to facilitate test packet checking
	RemoteDestIP2Cost  map[string]uint32    // store min cost of each route
	RemoteDestIP2SrcIP map[string]string    // store srcIP of each route to check Split Horizon with Poisoned Reverse to set cost = 16
	RemoteDest2ExTime  map[string]time.Time // store expiration time of each route to Check Expiration time of a new route
}

func (rt *RoutingTable) Make(args []string, nodePktOpChan chan *proto.NodePktOp, nodeExChan chan *proto.NodeEx, nodeSegRevChan chan *proto.Segment) {
	// Initialize ID2Interface and node packet channel
	rt.ID2Interface = make(map[uint8]*link.LinkInterface)
	// the interfaces need nodePktOpChan to pass received bytes
	rt.InitializeInterfaces(args, nodePktOpChan)
	// Expiration time
	rt.NodeExChan = nodeExChan
	// Initialize Routes: each interface to itself
	rt.DestIP2Route = map[string]Route{}
	// TCP Packet
	rt.NodeSegRevChan = nodeSegRevChan
	// initialize local IP set
	rt.LocalIPSet = map[string]bool{}
	for _, li := range rt.ID2Interface {
		route := Route{
			Dest: li.IPLocal,
			Next: li.IPLocal,
			Cost: 0,
		}
		rt.DestIP2Route[route.Dest] = route
		rt.LocalIPSet[li.IPLocal] = true
	}
	// initialize map remote2cost
	rt.RemoteDestIP2Cost = map[string]uint32{}
	// initialize map remoteDest2src
	rt.RemoteDestIP2SrcIP = map[string]string{}
	// initialize map remoteDest2exTime
	rt.RemoteDest2ExTime = map[string]time.Time{}
}

func (rt *RoutingTable) InitializeInterfaces(args []string, nodePktOpChan chan *proto.NodePktOp) {
	inx := args[1]
	f, err := os.Open(inx)
	if err != nil {
		log.Fatalln(err)
	}
	r := bufio.NewReader(f)

	id := uint8(0)

	var udpPortLocal string
	// Open linkConn with first line
	bytes, _, err := r.ReadLine()
	if err != nil {
		log.Fatalln("ReadFirstLine", err)
	}
	eles := strings.Split(string(bytes), " ")
	localAddr, err := net.ResolveUDPAddr("udp", ToIPColonAddr(eles[0], eles[1]))
	if err != nil {
		log.Fatalln("Resolve UDPAddr", err)
	}
	linkConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Fatalln("ListenUDP", err)
	}
	// Initialize link Interface
	for {
		bytes, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln(err)
		}
		eles := strings.Split(string(bytes), " ")
		li := &link.LinkInterface{}
		// elements: udpIp, udpPortRemote, ipLocal, ipRemote
		//li.Make(udpIp, udpPortRemote, ipLocal, ipRemote, id, udpPortLocal)
		li.Make(eles[0], eles[1], eles[2], eles[3], id, udpPortLocal, linkConn, nodePktOpChan)
		fmt.Printf("%v: %v\n", id, eles[2])
		rt.ID2Interface[id] = li
		id++
	}
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
