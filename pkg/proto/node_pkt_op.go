package proto

type NodePktOp struct {
	OpType uint8
	ID     uint8
	// packet: bytes of body
	Bytes interface{}
	// dest IP (which can be used to send packet)
	DestIP string
	// Send Packet
	ProtoID int
	Msg     string
}

// Receive one packet
func NewNodePktOp(opType, id uint8, bytes interface{}, destIP string, protoID int, msg string) *NodePktOp {
	nodePktOp := &NodePktOp{
		OpType:  opType,
		ID:      id,
		Bytes:   bytes,
		DestIP:  destIP,
		ProtoID: protoID,
		Msg:     msg,
	}
	return nodePktOp
}
