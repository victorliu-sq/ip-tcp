package proto

import (
	"log"
	"net"

	"github.com/google/netstack/tcpip/header"
	"golang.org/x/net/ipv4"
)

type PktTCP struct {
	Header *ipv4.Header
	Body   []byte
}

func NewPktTCP(IPSrc, IPDest string, msg []byte, ttl int) *PktTCP {
	body := msg
	header := NewTCPPktHeader(IPSrc, IPDest, len(body), ttl)
	// fmt.Println("length of body is:", len(body))
	tcpPkt := &PktTCP{
		Header: header,
		Body:   body,
	}
	headerBytes, err := tcpPkt.Header.Marshal()
	if err != nil {
		log.Fatalln("Error marshalling header:  ", err)
	}
	tcpPkt.Header.Checksum = int(ComputeChecksum(headerBytes))
	return tcpPkt
}

func (tcpPkt *PktTCP) Marshal() []byte {
	bytes, err := tcpPkt.Header.Marshal()
	if err != nil {
		log.Fatalln("Header Marshal Error", err)
	}
	bytes = append(bytes, tcpPkt.Body...)
	// fmt.Printf("Total length of rip is %v bytes\n", len(bytes))
	return bytes
}

func UnmarshalPktTCP(bytes []byte) *PktTCP {
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln(err)
	}
	body := bytes[20:]

	tcpPkt := &PktTCP{
		Header: header,
		Body:   body,
	}
	return tcpPkt
}

// *******************************************************************
// Test Header
func NewTCPPktHeader(IPSrc, IPDest string, bodyLen, ttl int) *ipv4.Header {
	return &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 20 + bodyLen,
		Flags:    header.IPv4FlagDontFragment,
		FragOff:  0,
		TTL:      ttl,
		Protocol: 6,
		Checksum: 0,
		Src:      net.ParseIP(IPSrc),
		Dst:      net.ParseIP(IPDest),
		Options:  []byte{},
	}
}
