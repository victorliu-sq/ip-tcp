package proto

import "time"

// API of node
const (
	// Command Line Interface
	CLI_SETUP          = uint8(0)
	CLI_SETDOWN        = uint8(1)
	CLI_QUIT           = uint8(2)
	CLI_LI             = uint8(3)
	CLI_LIFILE         = uint8(5)
	CLI_LR             = uint8(4)
	CLI_LRFILE         = uint8(6)
	CLI_LS             = uint8(7)
	CLI_LSFILE         = uint8(8)
	CLI_CREATELISTENER = uint8(9)
	CLI_CREATECONN     = uint8(10)
	CLI_SENDSEGMENT    = uint8(11)
	CLI_RECVSEGMENT    = uint8(12)
	CLI_BLOCKCLI       = uint8(13)
	CLI_UNBLOCKCLI     = uint8(14)
	CLI_CLOSE          = uint8(15)
	CLI_DELETECONN     = uint8(16)
	// network pass Packet to link
	MESSAGE_BCRIPREQ  = uint8(20)
	MESSAGE_BCRIPRESP = uint8(21)
	// Remote Route Expiration
	MESSAGE_ROUTEEX = uint8(22)
	// Send Packet to Link
	MESSAGE_SENDPKT = uint8(23)
	// Link pass packet back to network
	MESSAGE_REVPKT = uint8(24)

	PROTOCOL_RIP        = 200
	PROTOCOL_TESTPACKET = 0
	PROTOCOL_TCP        = 6

	LISTENER  = "LISTENER"
	LISTEN    = "LISTEN"
	SYN_SENT  = "SYN_SENT"
	SYN_RCVD  = "SYN_RCVD"
	ESTABLISH = "ESTABLISH"
	CLOSEWAIT = "CLOSE_W"
	FINWAIT1  = "FIN_W1"
	FINWAIT2  = "FIN_W2"
	TIMEWAIT  = "TIME_W"
	CLOSING   = "CLOSING"
	LASTACK   = "LAST_ACK"
	// the first port we allocate for conn
	FIRST_PORT = 0

	// DEFAULTPACKETMTU = 1400
	DEFAULTPACKETMTU = 20 + 20 + 3
	DEFAULTIPHDRLEN  = 20
	DEFAULTTCPHDRLEN = 20

	// Send Buffer
	// BUFFER_SIZE = 1 << 16
	BUFFER_SIZE        = uint32(10)
	SND_BUFFER_SIZE    = uint32(10)
	RCV_BUFFER_SIZE    = uint32(5)
	DEFAULT_DATAOFFSET = 20
	MAXCONNUM          = uint16(65535)

	RetranInterval = 300 * time.Millisecond
)

const TestString = "123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz"
