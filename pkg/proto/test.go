package proto

import (
	"log"
	"net"

	"golang.org/x/net/ipv4"
)

type PktTest struct {
	Header *ipv4.Header
	Body   []byte
}

func NewPktTest(IPSrc, IPDest, msg string, ttl int) *PktTest {
	body := []byte(msg)
	header := NewTestHeader(IPSrc, IPDest, len(body), ttl)
	test := &PktTest{
		Header: header,
		Body:   body,
	}
	headerBytes, err := test.Header.Marshal()
	if err != nil {
		log.Fatalln("Error marshalling header:  ", err)
	}
	test.Header.Checksum = int(ComputeChecksum(headerBytes))
	return test
}

func (test *PktTest) Marshal() []byte {
	bytes, err := test.Header.Marshal()
	if err != nil {
		log.Fatalln("Header Marshal Error", err)
	}
	bytes = append(bytes, test.Body...)
	// fmt.Printf("Total length of rip is %v bytes\n", len(bytes))
	return bytes
}

func UnmarshalPktTest(bytes []byte) *PktTest {
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln(err)
	}
	body := bytes[20:]

	test := &PktTest{
		Header: header,
		Body:   body,
	}
	return test
}

// *******************************************************************
// Test Header
func NewTestHeader(IPSrc, IPDest string, bodyLen, ttl int) *ipv4.Header {
	return &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 20 + bodyLen,
		Flags:    0,
		FragOff:  0,
		TTL:      ttl,
		Protocol: 0,
		Checksum: 0,
		Src:      net.ParseIP(IPSrc),
		Dst:      net.ParseIP(IPDest),
		Options:  []byte{},
	}
}
