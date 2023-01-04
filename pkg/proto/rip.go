package proto

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/google/netstack/tcpip/header"
	"golang.org/x/net/ipv4"
)

type PktRIP struct {
	Header *ipv4.Header
	Body   *RIPBody
}

func NewPktRIP(IPLocal, IPRemote string, commandType uint16, entries []Entry) *PktRIP {
	rip := &PktRIP{}
	rip.Body = NewRIPBody(entries, commandType)
	rip.Header = NewRIPHeader(IPLocal, IPRemote, len(rip.Body.Marshal()))
	headerBytes, err := rip.Header.Marshal()
	if err != nil {
		log.Fatalln("Error marshalling header:  ", err)
	}
	rip.Header.Checksum = int(ComputeChecksum(headerBytes))
	return rip
}

func (rip *PktRIP) Marshal() []byte {
	bytes, err := rip.Header.Marshal()
	// num of bytes in header is 20 bytes
	// fmt.Printf("num of bytes of Header is %v\n", len(bytes))
	if err != nil {
		log.Fatalln("Header Marshal Error", err)
	}
	bytes = append(bytes, rip.Body.Marshal()...)
	// fmt.Printf("Total length of rip is %v bytes\n", len(bytes))
	return bytes
}

func UnmarshalRIPResp(bytes []byte) *PktRIP {
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln(err)
	}
	body := UnmarshalRIPBody(bytes[20:])
	rip := &PktRIP{
		Header: header,
		Body:   body,
	}
	return rip
}

// ************************************************************************
// RIP Header
func NewRIPHeader(IPLocal, IPRemote string, bodyLen int) *ipv4.Header {
	header := &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 20 + bodyLen,
		Flags:    0,
		FragOff:  0,
		TTL:      16,
		Protocol: 200,
		Checksum: 0,
		Src:      net.ParseIP(IPLocal),
		Dst:      net.ParseIP(IPRemote),
		Options:  []byte{},
	}
	return header
}

// ************************************************************************
// RIP Body
type RIPBody struct {
	// command + num_entries = 4 bytes
	Command     uint16
	Num_Entries uint16
	// one entry = 12 bytes
	Entries []Entry
}

func NewRIPBody(entries []Entry, commandType uint16) *RIPBody {
	body := &RIPBody{
		Command:     commandType,
		Num_Entries: uint16(len(entries)),
		Entries:     entries,
	}
	return body
}

func (body *RIPBody) Marshal() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, body.Command)
	if err != nil {
		log.Fatalln(err)
	}
	err = binary.Write(buf, binary.BigEndian, body.Num_Entries)
	if err != nil {
		log.Fatalln(err)
	}
	bytes := buf.Bytes()
	// fmt.Printf("Length of body [Command + Num_Entries] is %v bytes\n", len(bytes))
	for _, entry := range body.Entries {
		bytes = append(bytes, entry.Marshal()...)
		// fmt.Printf("Length of body [Command + Num_Entries] + entry is %v bytes\n", len(bytes))
	}
	// fmt.Printf("Length of body is %v bytes\n", len(bytes))
	return bytes
}

func UnmarshalRIPBody(bytes []byte) *RIPBody {
	command := uint16(binary.BigEndian.Uint16(bytes[:2]))
	num_entries := uint16(binary.BigEndian.Uint16(bytes[2:4]))
	entries := []Entry{}
	for i := 0; i < int(num_entries); i++ {
		start, end := 4+i*12, 4+(i+1)*12
		entry := UnmarshalEntry(bytes[start:end])
		entries = append(entries, entry)
	}
	body := &RIPBody{
		Command:     command,
		Num_Entries: num_entries,
		Entries:     entries,
	}
	return body
}

func ComputeChecksum(b []byte) uint16 {
	checksum := header.Checksum(b, 0)
	// Invert the checksum value.  Why is this necessary?
	// The checksum function in the library we're using seems
	// to have been built to plug into some other software that expects
	// to receive the complement of this value.
	// The reasons for this are unclear to me at the moment, but for now
	// take my word for it.  =)
	checksumInv := checksum ^ 0xffff
	return checksumInv
}

// ************************************************************************
// RIP Entry
type Entry struct {
	// 12 byte
	Cost    uint32
	Address uint32
	Mask    uint32
}

func NewEntry(cost uint32, destAddr string) Entry {
	entry := Entry{
		Cost:    cost,
		Address: str2ipv4Num(destAddr),
		Mask:    1<<32 - 1,
	}
	return entry
}

func (entry Entry) Marshal() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, entry.Cost)
	if err != nil {
		log.Fatalln(err)
	}
	err = binary.Write(buf, binary.BigEndian, entry.Address)
	if err != nil {
		log.Fatalln(err)
	}
	err = binary.Write(buf, binary.BigEndian, entry.Mask)
	if err != nil {
		log.Fatalln(err)
	}
	bytes := buf.Bytes()
	return bytes
}

func UnmarshalEntry(bytes []byte) Entry {
	cost := uint32(binary.BigEndian.Uint32(bytes[:4]))
	address := uint32(binary.BigEndian.Uint32(bytes[4:8]))
	mask := uint32(binary.BigEndian.Uint32(bytes[8:]))
	entry := Entry{
		Cost:    cost,
		Address: address,
		Mask:    mask,
	}
	return entry
}

func str2ipv4Num(addr string) uint32 {
	numStrs := strings.Split(addr, ".")
	res := uint32(0)
	for _, numStr := range numStrs {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Fatalln(err)
		}
		res = res<<8 + uint32(num)
		// fmt.Println(res)
	}
	return res
}
