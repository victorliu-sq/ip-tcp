package transport

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"tcpip/pkg/proto"
)

// Writer to Write bytes into SND
func (conn *VTCPConn) FileWriter(Fd *os.File) {
	fd := Fd
	reader := bufio.NewReader(fd)
	for {
		content := make([]byte, proto.SND_BUFFER_SIZE)
		_, err := reader.Read(content)
		if err == io.EOF {
			DPrintf("All bytes In File has been written into SND buffer")
			break
		}
		conn.Writer(content)
	}
}

func (conn *VTCPConn) IsAllAcked() bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	fmt.Println("Wait", conn.snd.UNA, conn.snd.LBW)
	return conn.snd.UNA >= conn.snd.LBW
}

// Reader to read bytes from RCV
func (conn *VTCPConn) FileReader(Fd *os.File) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	fd := Fd
	cur := uint32(0)
	res := []byte{}
	for conn.State == proto.ESTABLISH {
		if !conn.rcv.IsEmpty() {
			bytes, bnum := conn.rcv.ReadFromBuffer(proto.RCV_BUFFER_SIZE)
			cur += bnum
			// fd.Write(bytes)
			fd.Write(bytes)
			res = append(res, bytes...)
			fmt.Println("************************************************")
			fmt.Println(cur)
			fmt.Println("************************************************")
			conn.rcv.PrintRCV()
		} else {
			fmt.Println("Read Sleep")
			conn.rcond.Wait()
			fmt.Println("Read Wake up")
		}
	}
	DPrintf("*******************Finish Receiving*******************")
	// if string(res) == proto.TestString {
	// 	println("************************************************")
	// 	fmt.Println("Woww!!!!!!!!")
	// 	println("************************************************")
	// }
	fd.Close()
}

func (conn *VTCPConn) IsFINRcvd() bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	// fmt.Println("Wait")
	return conn.State == proto.CLOSEWAIT
}
