package proto

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/google/netstack/tcpip/header"
	"golang.org/x/net/ipv4"
)

const (
	IpHeaderLen          = ipv4.HeaderLen
	TcpHeaderLen         = header.TCPMinimumSize
	TcpPsdueoHeaderLen   = 96
	IpProtoTcp           = header.TCPProtocolNumber
	MaxVirtualPacketSize = 1400
)

type Segment struct {
	IPhdr   *ipv4.Header
	TCPhdr  *header.TCPFields
	Payload []byte
}

func (segment *Segment) FormTuple() string {
	remotePort := segment.TCPhdr.SrcPort
	localPort := segment.TCPhdr.DstPort
	remoteAddr := segment.IPhdr.Src.String()
	localAddr := segment.IPhdr.Dst.String()
	return fmt.Sprintf("%v:%v:%v:%v", remoteAddr, remotePort, localAddr, localPort)
}

func NewSegment(IPSrc, IPDest string, tcpHdr *header.TCPFields, msg []byte) *Segment {
	body := []byte(msg)
	IPhdr := NewTCPPktHeader(IPSrc, IPDest, len(body), 0)
	seg := &Segment{
		IPhdr:   IPhdr,
		TCPhdr:  tcpHdr,
		Payload: msg,
	}
	checksum := ComputeTCPChecksum(tcpHdr, seg.IPhdr.Src, seg.IPhdr.Dst, body)
	tcpHdr.Checksum = checksum
	return seg
}

func UnMarshalSegment(hdr *ipv4.Header, bytes []byte) (*Segment, error) {
	tcpHeaderAndData := bytes[hdr.Len:hdr.TotalLen]
	tcpHdr := ParseTCPHeader(tcpHeaderAndData)
	tcpPayload := tcpHeaderAndData[tcpHdr.DataOffset:]

	tcpChecksumFromHeader := tcpHdr.Checksum // Save original
	tcpHdr.Checksum = 0
	tcpComputedChecksum := ComputeTCPChecksum(&tcpHdr, hdr.Src, hdr.Dst, tcpPayload)
	if tcpComputedChecksum != tcpChecksumFromHeader {
		return nil, fmt.Errorf("wrong tcp checksum")
	}

	segment := &Segment{
		IPhdr:   hdr,
		TCPhdr:  &tcpHdr,
		Payload: tcpPayload,
	}
	return segment, nil
}

func ParseTCPHeader(b []byte) header.TCPFields {
	td := header.TCP(b)
	return header.TCPFields{
		SrcPort:    td.SourcePort(),
		DstPort:    td.DestinationPort(),
		SeqNum:     td.SequenceNumber(),
		AckNum:     td.AckNumber(),
		DataOffset: td.DataOffset(),
		Flags:      td.Flags(),
		WindowSize: td.WindowSize(),
		Checksum:   td.Checksum(),
	}
}

func ComputeTCPChecksum(tcpHdr *header.TCPFields,
	sourceIP net.IP, destIP net.IP, payload []byte) uint16 {

	// Fill in the pseudo header
	pseudoHeaderBytes := make([]byte, 0, TcpPsdueoHeaderLen)
	pseudoHeaderBytes = append(pseudoHeaderBytes, sourceIP...) // 0..3
	pseudoHeaderBytes = append(pseudoHeaderBytes, destIP...)   // 4..7
	pseudoHeaderBytes[8] = 0
	pseudoHeaderBytes[9] = uint8(IpProtoTcp)

	totalLength := TcpHeaderLen + len(payload)
	binary.BigEndian.PutUint16(pseudoHeaderBytes[10:12], uint16(totalLength))

	// Turn the TcpFields struct into a byte array
	headerBytes := header.TCP(make([]byte, TcpHeaderLen))
	headerBytes.Encode(tcpHdr)

	// Compute the checksum for each individual part and combine To combine the
	// checksums, we leverage the "initial value" argument of the netstack's
	// checksum package to carry over the value from the previous part
	pseudoHeaderChecksum := header.Checksum(pseudoHeaderBytes, 0)
	headerChecksum := header.Checksum(headerBytes, pseudoHeaderChecksum)
	fullChecksum := header.Checksum(payload, headerChecksum)

	// Return the inverse of the computed value,
	// which seems to be the convention of the checksum algorithm
	// in the netstack package's implementation
	return fullChecksum ^ 0xffff
}

func PrintHex(bytes []byte) {
	fmt.Printf("[")
	for _, b := range bytes {
		fmt.Printf("%x ", b)
	}
	fmt.Printf("]")
	fmt.Println()
}
