package proto

import "os"

type NodeCLI struct {
	CLIType uint8
	ID      uint8
	// packet: bytes of body
	Bytes []byte
	// dest IP (which can be used to send packet)
	DestIP   string
	DestPort uint16
	// Send Packet
	ProtoID  int
	Msg      string
	Filename string
	Val16    uint16
	Val32    uint32
	Fd       *os.File
}

func NewNodeCLI(cliType, id uint8, bytes []byte, destIP string, destPort uint16, protoID int, msg string, filename string) *NodeCLI {
	nodeCLI := &NodeCLI{
		CLIType:  cliType,
		ID:       id,
		Bytes:    bytes,
		DestIP:   destIP,
		DestPort: destPort,
		ProtoID:  protoID,
		Msg:      msg,
		Filename: filename,
	}
	return nodeCLI
}
