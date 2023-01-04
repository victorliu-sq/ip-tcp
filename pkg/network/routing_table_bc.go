package network

import "tcpip/pkg/proto"

func (rt *RoutingTable) BroadcastRIPReq() {
	// fmt.Println("Try to broadcast RIP Req")
	for _, li := range rt.ID2Interface {
		if !li.IsUp() {
			continue
		}
		entries := []proto.Entry{}
		rip := proto.NewPktRIP(li.IPLocal, li.IPRemote, 1, entries)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}

func (rt *RoutingTable) BroadcastRIPResp() {
	// fmt.Println("Try to broadcast RIP Resp")
	for _, li := range rt.ID2Interface {
		if !li.IsUp() {
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
	}
}
