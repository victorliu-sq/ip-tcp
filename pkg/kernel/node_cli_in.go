package kernel

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"tcpip/pkg/proto"
	"time"
)

// Scan NodeCLI
/*
	When to output '> ' ?
	(1) After sending a cli to chan and  handling the cli, we can output a '>'
	(2) If the command is invalid and we cannot send it to chan, we output a '>' after error msg
*/

func (node *Node) ScanClI() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			// if node.blockCLI {
			// 	continue
			// }
			ws := strings.Split(line, " ")
			// fmt.Println(ws, len(ws), ws[0])
			if (len(ws) == 1 || len(ws) == 2) && ws[0] == "li" {
				if len(ws) == 1 {
					cli := proto.NewNodeCLI(proto.CLI_LI, 0, []byte{}, "", 0, 0, "", "")
					node.NodeCLIChan <- cli
				} else {
					// print li to a file
					cli := proto.NewNodeCLI(proto.CLI_LIFILE, 0, []byte{}, "", 0, 0, "", ws[1])
					node.NodeCLIChan <- cli
				}
			} else if (len(ws) == 1 || len(ws) == 2) && ws[0] == "lr" {
				if len(ws) == 1 {
					cli := proto.NewNodeCLI(proto.CLI_LR, 0, []byte{}, "", 0, 0, "", "")
					node.NodeCLIChan <- cli
				} else {
					// print lr to a file
					cli := proto.NewNodeCLI(proto.CLI_LRFILE, 0, []byte{}, "", 0, 0, "", ws[1])
					node.NodeCLIChan <- cli
				}
			} else if len(ws) == 2 && ws[0] == "up" {
				id, err := strconv.Atoi(ws[1])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", ws[1])
					continue
				}
				if id >= len(node.RT.ID2Interface) {
					fmt.Printf("interface %v does not exist\n> ", id)
					continue
				}
				// open link
				cli := proto.NewNodeCLI(uint8(proto.CLI_SETUP), uint8(id), []byte{}, "", 0, 0, "", "")
				node.NodeCLIChan <- cli
			} else if len(ws) == 2 && ws[0] == "down" {
				id, err := strconv.Atoi(ws[1])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", ws[1])
					continue
				}
				if id >= len(node.RT.ID2Interface) {
					fmt.Printf("interface %v does not exist\n> ", id)
					continue
				}
				// close link
				cli := proto.NewNodeCLI(uint8(proto.CLI_SETDOWN), uint8(id), []byte{}, "", 0, 0, "", "")
				node.NodeCLIChan <- cli
			} else if len(ws) == 1 && ws[0] == "q" {
				cli := proto.NewNodeCLI(proto.CLI_QUIT, 0, []byte{}, "", 0, 0, "", "")
				node.NodeCLIChan <- cli
			} else if len(ws) >= 4 && ws[0] == "send" {
				destIP := net.ParseIP(ws[1]).String()
				protoID, err := strconv.Atoi(ws[2])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", ws[1])
					continue
				}
				msg := line[len(ws[0])+len(ws[1])+len(ws[2])+3:]
				// fmt.Println(msg)
				// cli := proto.NewNodeCLI(proto.MESSAGE_SENDPKT, 0, []byte{}, destIP, protoID, msg)
				cli := proto.NewNodeCLI(proto.MESSAGE_SENDPKT, 0, []byte{}, destIP, 0, protoID, msg, "")
				node.NodeCLIChan <- cli
			} else if len(ws) == 2 && ws[0] == "a" { //a port
				portNum, err := strconv.Atoi(ws[1])
				if err != nil {
					fmt.Printf("strconv.Atoi: parsing %v: invalid syntax\n> ", ws[1])
					continue
				}
				cli := &proto.NodeCLI{CLIType: proto.CLI_CREATELISTENER, Val16: uint16(portNum)}
				node.NodeCLIChan <- cli
			} else if (len(ws) == 1 || len(ws) == 2) && ws[0] == "ls" {
				if len(ws) == 1 {
					cli := proto.NewNodeCLI(proto.CLI_LS, 0, []byte{}, "", 0, 0, "", "")
					node.NodeCLIChan <- cli
				} else {
					// print lr to a file
					cli := proto.NewNodeCLI(proto.CLI_LSFILE, 0, []byte{}, "", 0, 0, "", ws[1])
					node.NodeCLIChan <- cli
				}
			} else if len(ws) == 3 && ws[0] == "c" {
				remoteAddr := ws[1]
				remotePortS := ws[2]
				remotePort, err := strconv.Atoi(remotePortS)
				if err != nil {
					continue
				}
				cli := proto.NewNodeCLI(proto.CLI_CREATECONN, 0, []byte{}, remoteAddr, uint16(remotePort), 0, "", "")
				node.NodeCLIChan <- cli
			} else if len(ws) == 3 && ws[0] == "s" {
				// id, err := strconv.Atoi(ws[1])
				// if err != nil {
				// 	continue
				// }
				// // content := []byte(proto.TestString)
				// content := []byte(ws[2])
				// cli := &proto.NodeCLI{CLIType: proto.CLI_SENDSEGMENT, Val16: uint16(id), Bytes: content}
				// node.NodeCLIChan <- cli
			} else if len(ws) == 4 && ws[0] == "r" {
				// socketId, err := strconv.Atoi(ws[1])
				// if err != nil {
				// 	continue
				// }
				// numBytes, err := strconv.Atoi(ws[2])
				// if err != nil {
				// 	continue
				// }
				// isBlock := []byte(ws[3])
				// cli := &proto.NodeCLI{CLIType: proto.CLI_RECVSEGMENT, Bytes: isBlock,
				// 	Val16: uint16(socketId), Val32: uint32(numBytes)}
				// node.NodeCLIChan <- cli
			} else if len(ws) == 2 && ws[0] == "cl" {
				// socketId, err := strconv.Atoi(ws[1])
				// if err != nil {
				// 	return
				// }
				// cli := &proto.NodeCLI{CLIType: proto.CLI_CLOSE, Val16: uint16(socketId)}
				// node.NodeCLIChan <- cli
			} else if len(ws) == 3 && ws[0] == "rf" {
				// path := ws[1]
				// fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0x666)
				// if err != nil {
				// 	fmt.Printf("%v\n", err)
				// 	continue
				// }
				// port, err := strconv.Atoi(ws[2])
				// if err != nil {
				// 	fmt.Printf("%v\n", err)
				// 	continue
				// }
				// ls := node.socketTable.OfferListener(uint16(port))
				// go node.NodeAcceptLoop(ls, true)
				// ls.CLIChan <- &proto.NodeCLI{Fd: fd}
			} else if len(ws) == 2 && ws[0] == "cl" {
				// socketId, err := strconv.Atoi(ws[1])
				// if err != nil {
				// 	continue
				// }
				// node.socketTable.DeleteSocket(uint16(socketId))
				// fmt.Printf("\n> ")
			} else if len(ws) == 4 && ws[0] == "sf" {
				// fd, err := os.OpenFile(ws[1], os.O_RDONLY, 0x666)
				// if err != nil {
				// 	fmt.Printf("%v\n", err)
				// 	continue
				// }
				// port, err := strconv.Atoi(ws[3])
				// if err != nil {
				// 	fmt.Printf("%v\n", err)
				// 	continue
				// }
				// cli := &proto.NodeCLI{DestIP: ws[2], DestPort: uint16(port)}
				// conn := node.HandleCreateConn(cli)
				// conn.Fd = fd
				// go conn.VSBufferWriteFile()

				//pull data from fd
				//wait for all the data is sent
			} else {
				fmt.Printf("Invalid command\n> ")
			}
		}
	}
}

// Send NodeBroadcast
func (node *Node) RIPRespDaemon() {
	for {
		cli := proto.NewNodeBC(proto.MESSAGE_BCRIPRESP, 0, []byte{}, "", 0, "")
		node.NodeBCChan <- cli
		time.Sleep(5 * time.Second)
	}
}

func (node *Node) RIPReqDaemon() {
	cli := proto.NewNodeBC(proto.MESSAGE_BCRIPREQ, 0, []byte{}, "", 0, "")
	node.NodeBCChan <- cli
}
