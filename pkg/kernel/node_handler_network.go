package kernel

import (
	"log"
	"tcpip/pkg/proto"

	"golang.org/x/net/ipv4"
)

// ***********************************************************************************
// Handle Receive Packet
func (node *Node) HandleReceivePacket(bytes []byte, destAddr string) {
	// check if  match can any port and the port is still alive
	// fmt.Println("Receive a packet")
	if !node.RT.CheckPktValidity(bytes, destAddr) {
		return
	}
	h, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln("Parse Header", err)
	}
	// HandleRIPResp or HandleTest
	switch h.Protocol {
	case proto.PROTOCOL_RIP:
		b := proto.UnmarshalRIPBody(bytes[20:])
		if b.Command == 1 {
			// fmt.Printf("Receive a RIP Req Packet from %v\n", destAddr)
			node.RT.HandleRIPReq(h.Src.String())
		} else {
			// fmt.Printf("Receive a RIP Resp Packet from %v\n", destAddr)
			node.RT.HandleRIPResp(bytes)
		}
	case proto.PROTOCOL_TESTPACKET:
		// fmt.Printf("Receive a TEST Packet from %v\n", destAddr)
		node.RT.ForwardTestPkt(bytes)
	case proto.PROTOCOL_TCP:
		// fmt.Printf("Receive a TCP Packet from %v\n", destAddr)
		node.RT.ForwardTCPPkt(h, bytes)
	}
}
