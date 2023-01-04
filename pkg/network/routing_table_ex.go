package network

import (
	"tcpip/pkg/proto"
	"time"
)

func (rt *RoutingTable) CheckRouteEx(destIP string) {
	if time.Now().After(rt.RemoteDest2ExTime[destIP]) {
		// fmt.Println(destIP, "Del 275")
		newRoute := rt.DestIP2Route[destIP]
		newRoute.Cost = 16
		rt.UpdateRoutesAndBroadcastTU(newRoute, destIP)
	}
}

// Send NodeEx
func (rt *RoutingTable) SendExTimeCLI(destIP string) {
	// sleep 12 second and check whether the time expires
	time.Sleep(13 * time.Second)
	cli := proto.NewNodeEx(proto.MESSAGE_ROUTEEX, 0, []byte{}, destIP, 0, "")
	rt.NodeExChan <- cli
}

// Update Expiration time of routes
func (rt *RoutingTable) UpdateExTime(destIP string) {
	rt.RemoteDest2ExTime[destIP] = time.Now().Add(12 * time.Second)
	go rt.SendExTimeCLI(destIP)
}
