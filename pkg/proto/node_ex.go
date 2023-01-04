package proto

type NodeEx struct {
	OpType uint8
	ID     uint8
	// packet: bytes of body
	Bytes []byte
	// dest IP (which can be used to send packet)
	DestIP string
	// Send Packet
	ProtoID int
	Msg     string
}

// Node Expiration
func NewNodeEx(opType, id uint8, bytes []byte, destIP string, protoID int, msg string) *NodeEx {
	nodeEx := &NodeEx{
		OpType:  opType,
		ID:      id,
		Bytes:   bytes,
		DestIP:  destIP,
		ProtoID: protoID,
		Msg:     msg,
	}
	return nodeEx
}
