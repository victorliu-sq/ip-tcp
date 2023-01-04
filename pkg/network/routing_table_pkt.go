package network

import (
	"fmt"
	"log"
	"strconv"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"

	"golang.org/x/net/ipv4"
)

// *****************************************************************************
// Send a packet
func (rt *RoutingTable) SendPacket(destIP string, msg string) {
	ttl := 16
	if route, ok := rt.DestIP2Route[destIP]; ok && route.Cost < 16 {
		// check if route.cost == inf => unreachable
		// Choose the link whose IPRemote == nextIP to send
		for _, li := range rt.ID2Interface {
			if li.IPRemote == route.Next {
				// fmt.Printf("Try to send a packet from %v to %v\n", li.IPLocal, destIP)
				test := proto.NewPktTest(li.IPLocal, destIP, msg, ttl-1)
				bytes := test.Marshal()
				// proto.PrintHex(bytes)
				li.SendPacket(bytes)
				return
			}
		}
	}
	fmt.Println("destIP does not exist")
}

func (rt *RoutingTable) SendTCPPacket(srcIP, destIP string, msg string) {
	ttl := 16
	if route, ok := rt.DestIP2Route[destIP]; ok && route.Cost < 16 {
		// check if route.cost == inf => unreachable
		// Choose the link whose IPRemote == nextIP to send
		for _, li := range rt.ID2Interface {
			if li.IPRemote == route.Next {
				// tcpPkt := proto.NewPktTCP(li.IPLocal, destIP, []byte(msg), ttl-1)
				tcpPkt := proto.NewPktTCP(srcIP, destIP, []byte(msg), ttl-1)
				bytes := tcpPkt.Marshal()
				// proto.PrintHex(bytes)
				li.SendPacket(bytes)
				return
			}
		}
	}
	fmt.Println("destIP does not exist")
}

// *****************************************************************************
// Receive a packet
func (rt *RoutingTable) CheckPktValidity(bytes []byte, destAddr string) bool {
	canMatch := false
	isAlive := false
	for _, li := range rt.ID2Interface {
		if destAddr == li.MACRemote {
			canMatch = true
			if li.IsUp() {
				isAlive = true
				break
			}
		}
	}
	if !canMatch || !isAlive {
		// link to is down but link from is still up
		// fmt.Printf("%v receive a msg from dead link\n", destAddr)
		return false
	}
	// check length of bytes
	if len(bytes) < 20 {
		// fmt.Println(len(bytes))
		return false
	}
	h, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln("Parse Header", err)
	}
	if h.TotalLen != len(bytes) {
		// fmt.Println(h.TotalLen, len(bytes))
		return false
	}
	// 1. Validity
	// (1) Is checksum in header valid
	prevChecksum := h.Checksum
	// we need toset checksum back to 0 before computing the current checksum
	h.Checksum = 0
	hBytes, err := h.Marshal()
	if err != nil {
		log.Fatalln(err)
	}
	curChecksum := int(proto.ComputeChecksum(hBytes))
	if prevChecksum != curChecksum {
		fmt.Println("Should be:", h.Checksum, ", Current:", curChecksum)
		fmt.Println("Receive:", h)
		return false
	}
	// (2) Check if TTL == 0
	if h.TTL == 0 {
		fmt.Println("No enough TTL")
		return false
	}
	return true
}

func (rt *RoutingTable) HandleRIPReq(srcIP string) {
	for _, li := range rt.ID2Interface {
		if !li.IsUp() || li.IPRemote != srcIP {
			continue
		}
		entries := []proto.Entry{}
		// For RIP resp, we need to load all valid entries into RIP body
		for _, route := range rt.DestIP2Route {
			// if route.next == src of route.dest -> ignore this route entry
			entry := proto.NewEntry(route.Cost, route.Dest)
			if srcIP, ok := rt.RemoteDestIP2SrcIP[route.Dest]; ok && srcIP == li.IPRemote {
				entry.Cost = 16
			}
			entries = append(entries, entry)
			// fmt.Println(entries)
		}
		rip := proto.NewPktRIP(li.IPLocal, li.IPRemote, 2, entries)
		// fmt.Println("Send", rip.Header)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
		break
	}
}

func (rt *RoutingTable) HandleRIPResp(bytes []byte) {
	rip := proto.UnmarshalRIPResp(bytes)
	num_entries := rip.Body.Num_Entries
	// fmt.Println(num_entries)
	for i := 0; i < int(num_entries); i++ {
		entry := rip.Body.Entries[i]
		// fmt.Println(entry)
		// 1. Expiration time
		// If the destIP is local IP of 1 interface, it will not expire
		destIP := ipv4Num2str(entry.Address)
		if _, ok := rt.LocalIPSet[destIP]; ok {
			continue
		}

		// 2. Route
		oldCost := rt.DestIP2Route[destIP].Cost
		oldNextAddr := rt.DestIP2Route[destIP].Dest
		var newCost uint32
		newNextAddr := rip.Header.Src.String()
		if entry.Cost == 16 {
			newCost = 16
		} else {
			newCost = entry.Cost + 1
		}
		newRoute := NewRoute(destIP, newNextAddr, newCost)
		// (1) If no existing route and cost != 16 , update
		if _, ok := rt.RemoteDestIP2Cost[destIP]; !ok {
			rt.UpdateRoutesAndBroadcastTU(newRoute, destIP)
			rt.UpdateExTime(destIP)
		}
		// (2) If newCost < oldCost, update
		if newCost < oldCost {
			rt.UpdateRoutesAndBroadcastTU(newRoute, destIP)
			rt.UpdateExTime(destIP)
		}
		// (3) If newCost > oldCost and newNextAddr == oldNextAddr, update
		// if cost == 16 and newNextAddr == oldNextAddr => expired
		if newCost > oldCost && newNextAddr == oldNextAddr {
			rt.UpdateRoutesAndBroadcastTU(newRoute, destIP)
			rt.UpdateExTime(destIP)
		}
		// cost == 16
		// Situation1: routes expires => solved in step3
		// Situation2: send back routes => in step4

		// (4) If newCost > oldCost and newNextAddr != oldNextAddr, ignore
		// if cost != 16 and newCost > old and newNextAddr != oldNextAddr => worse route => ignore
		// if cost == 16 and newCost > old and newNextAddr != oldNextAddr => send back / another path is dead => ignore
		if newCost > oldCost && newNextAddr != oldNextAddr {
			continue
		}
		// (5) If newCost == oldCost, reset the expire time (done)
		if newCost == oldCost {
			rt.UpdateExTime(destIP)
		}
	}
}

func ipv4Num2str(addr uint32) string {
	mask := 1<<8 - 1
	res := strconv.Itoa(int(addr) & mask)
	addr >>= 8
	for i := 0; i < 3; i++ {
		res = strconv.Itoa(int(addr)&mask) + "." + res
		addr >>= 8
	}
	return res
}

// ***********************************************************************************
// Handle Test Packet
func (rt *RoutingTable) ForwardTestPkt(bytes []byte) {
	test := proto.UnmarshalPktTest(bytes)
	srcIP := test.Header.Src.String()
	destIP := test.Header.Dst.String()
	msg := string(test.Body)
	ttl := test.Header.TTL
	fmt.Println("Get one test packet")
	// 2. Forwarding
	// (1) Does this packet belong to me?
	if _, ok := rt.LocalIPSet[destIP]; ok {
		fmt.Printf("---Node received packet!---\n")
		fmt.Printf("        source IP      : %v\n", srcIP)
		fmt.Printf("        destination IP : %v\n", destIP)
		fmt.Printf("        protocol       : %v\n", 0)
		fmt.Printf("        payload length : %v\n", len(msg))
		fmt.Printf("        payload        : %v\n", msg)
		fmt.Printf("----------------------------\n")
		return
	}
	// (2) Does packet match any route in the forwarding table?
	if route, ok := rt.DestIP2Route[destIP]; ok && route.Cost < 16 {
		// Choose the link whose IPRemote == nextIP to send
		for _, li := range rt.ID2Interface {
			if li.IPRemote == route.Next {
				fmt.Printf("Try to send a packet from %v to %v\n", li.IPLocal, destIP)
				test := proto.NewPktTest(srcIP, destIP, msg, ttl-1)
				bytes := test.Marshal()
				li.SendPacket(bytes)
				return
			}
		}
	}
	// (3) Does the router have next hop?
	fmt.Println("destIP does not exist")
}

// ***********************************************************************************
// Handle TCP Packet
func (rt *RoutingTable) ForwardTCPPkt(h *ipv4.Header, bytes []byte) {
	segment, err := proto.UnMarshalSegment(h, bytes)
	if err != nil {
		return
	}

	srcIP := h.Src.String()
	destIP := h.Dst.String()
	msg := bytes
	ttl := h.TTL
	// 2. Forwarding
	// (1) Does this packet belong to me?
	if _, ok := rt.LocalIPSet[destIP]; ok {
		rt.SegRevChan <- segment
		return
	}
	// (2) Does packet match any route in the forwarding table?
	if route, ok := rt.DestIP2Route[destIP]; ok && route.Cost < 16 {
		// Choose the link whose IPRemote == nextIP to send
		for _, li := range rt.ID2Interface {
			if li.IPRemote == route.Next {
				fmt.Printf("Try to send a packet from %v to %v\n", li.IPLocal, destIP)
				segment := proto.NewPktTCP(srcIP, destIP, msg, ttl-1)
				bytes := segment.Marshal()
				li.SendPacket(bytes)
				myDebug.Debugln("Forward one TCP packet")
				return
			}
		}
	}
	// (3) Does the router have next hop?
	fmt.Println("destIP does not exist")
}

// ***********************************************************************************
// Update routes
func (rt *RoutingTable) UpdateRoutesAndBroadcastTU(newRoute Route, destIP string) {
	// update routes
	rt.DestIP2Route[destIP] = newRoute
	// update the metadata
	rt.RemoteDestIP2Cost[destIP] = newRoute.Cost
	rt.RemoteDestIP2SrcIP[destIP] = newRoute.Next
	// Broadcast RIP Resp because of Triggered Updates
	entry := proto.NewEntry(newRoute.Cost, newRoute.Dest)
	rt.BroadcastRIPRespTU(entry)
}

// Send Triggered Updates
func (rt *RoutingTable) BroadcastRIPRespTU(entity proto.Entry) {
	for _, li := range rt.ID2Interface {
		if !li.IsUp() {
			continue
		}
		entries := []proto.Entry{}
		entries = append(entries, entity)
		rip := proto.NewPktRIP(li.IPLocal, li.IPRemote, 2, entries)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}

// ***********************************************************************************
// Find Src IP addr
func (rt *RoutingTable) FindSrcIPAddr(destIP string) string {
	if route, ok := rt.DestIP2Route[destIP]; ok {
		for _, li := range rt.ID2Interface {
			if li.IPRemote == route.Next {
				return li.IPLocal
			}
		}
	}
	return "no"
}
