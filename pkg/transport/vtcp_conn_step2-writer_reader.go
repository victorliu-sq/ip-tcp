package transport

import (
	"fmt"
	"tcpip/pkg/proto"
)

// Writer to Write bytes into SND
func (conn *VTCPConn) Writer(content []byte) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	total := uint32(len(content))
	for total > 0 {
		if !conn.snd.IsFull() {
			bnum := conn.snd.WriteIntoBuffer(content)
			total -= bnum
			content = content[bnum:]
			conn.scond.Signal()
			// Print current snd
			conn.snd.PrintSND()
		} else {
			fmt.Println("Write Sleep")
			conn.wcond.Wait()
			fmt.Println("Write Wake up")
		}
	}
}

// Reader
// Reader to Read bytes from RCV
func (conn *VTCPConn) Reader(total uint32) []byte {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	res := []byte{}
	for total > 0 {
		if !conn.rcv.IsEmpty() {
			bytes, bnum := conn.rcv.ReadFromBuffer(total)
			total -= bnum
			res = append(res, bytes...)
			fmt.Println("************************************************")
			fmt.Println("Read", bnum, "bytes")
			fmt.Println(string(res))
			fmt.Println("************************************************")
			conn.rcv.PrintRCV()
		} else {
			fmt.Println("Read Sleep")
			conn.rcond.Wait()
			fmt.Println("Read Wake up")
		}
	}
	DPrintf("*******************Finish Sending*******************")
	if string(res) == proto.TestString {
		println("************************************************")
		fmt.Println("Woww!!!!!!!!")
		println("************************************************")
	}
	return res
}
