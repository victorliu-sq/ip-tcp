package link

import (
	"fmt"
	"log"
	"net"
	"sync"
	"tcpip/pkg/proto"
)

type LinkInterface struct {
	Mu        sync.Mutex
	ID        uint8
	MACLocal  string
	MACRemote string
	IPLocal   string
	IPRemote  string
	Status    string
	// we use RemoteConn to send data to remote machine
	LinkConn *net.UDPConn
	// we use this channel to send packet back to node
	NodePktOpChan chan *proto.NodePktOp
}

func (li *LinkInterface) Make(udpIp, udpPortRemote, ipLocal, ipRemote string,
	id uint8, udpPortLocal string, linkConn *net.UDPConn, nodePktOpChan chan *proto.NodePktOp) {
	li.ID = id
	li.MACLocal = udpIp + ":" + udpPortLocal
	li.MACRemote = udpIp + ":" + udpPortRemote
	li.IPLocal = ipLocal
	li.IPRemote = ipRemote
	// Communication between layer and network
	li.NodePktOpChan = nodePktOpChan
	if li.IPLocal == "" {
		return
	}
	// LocalAddr
	// Setup RemoteConn
	remoteAddr, err := net.ResolveUDPAddr("udp", li.MACRemote)
	if err != nil {
		log.Fatalln(err)
	}
	localAddr, err := net.ResolveUDPAddr("udp", li.MACLocal)
	if err != nil {
		log.Fatalln(err)
	}
	li.MACRemote = remoteAddr.String()
	li.MACLocal = localAddr.String()
	// fmt.Println(li.MACLocal, li.MACRemote)
	li.LinkConn = linkConn
	if err != nil {
		log.Fatalln("Open LinkConn", err)
	}
	li.Status = "up"
	go li.ServeLink()
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
