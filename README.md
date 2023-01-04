# Latedays

We use 1 late day for this project



# Lab

Lab has been finished on time

You can check the first commit of file lab.md on the branch of jiaxin



# Node

The main idea of my node is to leverage channel to avoid data racing. To be specific, all operations of node including handling packets, updating routing table, and checking expired routes will be linearized by sending to channel before they are down.

There will be 4 channels in each node:

| channel       | functionality                                                |
| ------------- | ------------------------------------------------------------ |
| NodeBCChan    | Broadcast RIP messages(RIP requets, RIP respones and triggered updates) |
| NodeExChan    | Check whether the routes have expired (delete expired routes from routing table) |
| NodePktOpChan | Handle received bytes of packets passed from Link Interface (handle RIP packets and Test packets) |
| NodeCLIChan   | Handle CLI from user (lr, li, down, up)                      |

 The link interface will interact with node though NodePktOpChan so we need to pass it to link interface when initializing it.



# Routing Table

Each node has one routing table which can be used to send packets though link

| Metadata           | functionality                       |
| ------------------ | ----------------------------------- |
| RemoteDestIP2Cost  | store min cost of each route        |
| RemoteDestIP2SrcIP | store srcIP of each route           |
| RemoteDest2ExTime  | store expiration time of each route |

| Variables    | functionality         |
| ------------ | --------------------- |
| ID2Interface | store link Interfaces |
| DestIP2Route | store routes          |





# Link Interface

Each link interface includes the vitual link (UDP socket) to send/receive bytes to/from neighbors and metadata about IP address and MAC address of linked nodes.



The structure of Link Interface looks like this:

| Variable name | Variable Type         | Functionality                         |
| ------------- | --------------------- | ------------------------------------- |
| Mu            | sync.Mutex            | protect Status of link interface      |
| ID            | uint8                 | id of                                 |
| MACLocal      | string                | MAC address of local node             |
| MACRemote     | string                | MAC address of remote node            |
| IPLocal       | string                | local IP address of the interface     |
| IPRemote      | string                | remote IP address of the interface    |
| Status        | string                | mark this link interface is up / down |
| LinkConn      | *net.UDPConn          | read / write bytes from/into the link |
| NodePktOpChan | chan *proto.NodePktOp | send bytes of packet back to its node |

The link interface will listen on the port number in the Inx file and send bytes with the same port number.

When link interface receive some bytes, it will check whether the sender has the same remoteAddr as its. The link interface will abandon the data if the addresses are different. Besides, if status of remote link is "down", the data will be refused either 





# RIP Packets

There are 5 cases of sending RIP packets:

- When the node comes online, broadcast RIP request packets

- After coming online, broadcast RIP response packets every 5 seconds periodically 

  - If the route entry is sent back to source node, set the cost of route entry to 16(infinity)

- If 1 route entry gets updated, broadcast triggered updates to neighbors

- If 1 route entry expires, broadcast triggered udpate( cost of route entry == 16) to neighbors

- If status of 1 link is set to "down", broadcast triggered udpate( cost of route entry == 16) to neighbors

RIP messages will be marshalled to bytes then sent through the link which has corresponding next hop. The remote link will pass the bytes into NodeCLIChan then the node running upon this link will handle this RIP packet like this: 

- Check Validity:
  - Check whether the checksum in header is valid
  - Check if TTL == 0
- Check metadata of RIP Packet:
  - if destIP is local IP of current node, it will not expire
  - if destIP does not exist in routing table of current node, add the route to routing table and reset its expiration time
  - if newCost < oldCost, update cost of this route to smaller cost and reset its expiration time
  - if newCost > oldCost and newNextIPAddress == oldNextIPAddress, update cost of this route to larger one and reset its expiration time. If newCost == 16, delete newly added route from routing table because this is a triggered update caused by expiration of the route
  - if newCost > oldCost and newNextIPAddress != oldNextIPAddress, ignore this route entry
  - If newCost == oldCost, reset the expiration time



# Test Packets

The node receiving commands from user will try to send a test packet to destIP through link interface whose next hop IP fits the routing table. Other nodes receiving bytes of test packets will handle the packets like this: 

- Check Validity:
  - Check whether the checksum in header is valid
  - Check if TTL == 0
- Check metadata of Test Packet:
  - If destIP is one of local IP addresses of current node, print out this test msg
  - If destIP matches any route in the routing table, send it though the corresponding link interface
  - If destIP does not any route, stop routing

