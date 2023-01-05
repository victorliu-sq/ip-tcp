package network

import "tcpip/pkg/proto"

// ***********************************************************************************
// Find Src IP addr
func (rt *RoutingTable) FindSrcIPAddr(destIP string) (string, bool) {
	if route, ok := rt.DestIP2Route[destIP]; ok {
		for _, li := range rt.ID2Interface {
			if li.IPRemote == route.Next {
				return li.IPLocal, true
			}
		}
	}
	return "", false
}

func (rt *RoutingTable) SendToNodeSegChannel(segment *proto.Segment) {
	rt.NodeSegRevChan <- segment
}
