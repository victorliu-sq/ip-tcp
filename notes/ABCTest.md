# Command



| Command                          | Description                                                  |
| :------------------------------- | :----------------------------------------------------------- |
| `interfaces`, `li`               | Prints information about each interface, one per line.       |
| `interfaces <file>`, `li <file>` | Print information about each interface, one per line, to the destination file. Overrides the file if it exists. |
| `routes`, `lr`                   | Print information about the route to each known destination, one per line. |
| `routes <file>`, `lr <file>`     | Print information about the route to each known destination, one per line, to the destination file. Overwrites the file if it exists. |
| `up <integer>`                   | Bring an interface with ID `<integer>` “up” (it must be an existing interface, probably one you brought down). |
| `send <vip> <proto> <string>`    | Send an IP packet with protocol `<proto>` (an integer) to the virtual IP address `<vip>` (dotted quad notation). The payload is simply the characters of `<string>` (as in Snowcast, do not null-terminate this). |
| `q`                              | Quit the node by cleaning up used resources.                 |





# Test

start node

```shell
reference node + inx

./tools/ref_node ./nets/routeAggregation/tree/A.lnx
./tools/ref_node ./nets/routeAggregation/tree/B.lnx
./tools/ref_node ./nets/routeAggregation/tree/C.lnx

./node ./nets/routeAggregation/tree/A.lnx
./node ./nets/routeAggregation/tree/B.lnx
./node ./nets/routeAggregation/tree/C.lnx
```

print all link interfaces

```shell
li
```

test down/up

```shell
up 0
li
down 0
li
up 0
```



print all routers

```shell
# my nodeA
# my nodeB
# ref nodeC

lr
```



send packets

```shell
send <vip> <proto> <string>  
#A send to A
send 10.0.0.1 0 Hello from A

#A send to B
send 10.0.0.14 0 Hello from A

#A send to C
send 10.0.0.10 0 Hello from A
```



kill process

```shell
sudo install lsof
#(get process who's listening on 17001)
sudo lsof -i -P -n | grep 17001 

kill + pid
```



cound lines in a dir

```shell
git ls-files | xargs cat | wc -l
```



# Link

## Close

if a linkInterace is down, the node should not receive any packets from it.

node -> move the disabled route from (1) LocalIPSet (2) DestIP2Route map[string]Route

​		-> move all routes whose next IP is the remote IP addr of the disabled route

Link -> change status



## Open

if a linkInterace is up, the node should not receive any packets from it.

node -> add the enabled route from (1) LocalIPSet (2) DestIP2Route map[**string**]Route (3) all metadata of remoteDestIP

Link -> change status, start if **serveLinkli**



# Packet

## Send





## Receive

When we receive an RIP msg

- Send one command line interface of RIPHandle to channel







# RIP

## RIP Request

### struct

Header

| element  |                        | functionality                                         |
| -------- | ---------------------- | ----------------------------------------------------- |
| Protocol | 200                    |                                                       |
| Len      | 120                    | avoid err "Header Marshal Error header too short"     |
| Src      | IP of local interface  | next hop for new route, to avoid go back              |
| Dest     | IP of remote interface | target node can use it to find which link to response |



Body

| Element       | Type    | value                                                        |
| ------------- | ------- | ------------------------------------------------------------ |
| command       | uint16  | `1` for a request of routing information, and `2` for a response |
| num_entries   | uint16  | 0                                                            |
| entries       | []Entry | empty                                                        |
| Entry.cost    | uint32  |                                                              |
| Entry.address | uint32  |                                                              |
| Entry.mask    | uint32  |                                                              |



### Send RIP Request

Periodically broadcast all RIP requests to neighbor nodes



### Receve RIP Request

Once a node receives a RIP Request, it will send a CLI of RIPReq to CLIChannel

- node will send RIP Response back to Src IP



## RIP Response

### struct

Header

| element  |                       | functionality                                     |
| -------- | --------------------- | ------------------------------------------------- |
| Protocol | 200                   |                                                   |
| Len      | 120                   | avoid err "Header Marshal Error header too short" |
| Src      | IP of local interface | next hop for new route                            |



Body

| Element       | Type    | value                                                        |
| ------------- | ------- | ------------------------------------------------------------ |
| command       | uint16  | `1` for a request of routing information, and `2` for a response |
| num_entries   | uint16  | len(entries)                                                 |
| entries       | []Entry | all valid entries that do not go back to source of IP        |
| Entry.cost    | uint32  | current route entry, if Sending backing, set it to infinity(16) |
| Entry.address | uint32  | Dest of current route entry                                  |
| Entry.mask    | uint32  | 1 << 32 - 1                                                  |



### min_cost of a routing entry

| metadata          | type              |                                     |
| ----------------- | ----------------- | ----------------------------------- |
| RemoteDestIP2Cost | map[string]uint32 | record min cost of remote dest addr |

When we receive a RIP msg, check all of its route entries, new cost = entry.cost + 1

- if dest addr exits in RemoteDestIP2Cost && newCost >= oldCost: ignore it
- else: create a newroute and store it





### Split Horizon with Poisoned Reverse

| metadata           | type              |                                 |
| ------------------ | ----------------- | ------------------------------- |
| RemoteDestIP2SrcIP | map[string]string | record src ip of remote dest ip |

When we are sending out RIP

- if next ip of this dest ip is src ip, set its cost to infinity(16)
- else put the entry into body of the packet





### Expiration of a routing entry

| metadata              | type                 | functionality                       |
| --------------------- | -------------------- | ----------------------------------- |
| RemoteDest2ExpireTime | map[string]time.Time | record when the route should expire |

When we receive a RIP msg:

- update expration time of remote dest IP to time.Now().Add(12 * time.Second) 
- start a goroutine to check expiration after 12 second



In terms of Expiration Goroutine:

- Send one CLI of Check Expiration to channel after 12 seconds
- CLI will check the expiration time of that dest IP:
  - if time.Now().After(ExpirationTime of dest IP)
    - delete destIP from RemoteDest2ExTime
    - delete destIP from DestIP2Route
    - delete destIP from RemoteDestIP2Cost
    - delete destIP from RemoteDestIP2SrcIP
  - else: do nothing



### Triggered updates

Sender:

After handling a RIP response, if the cost of Remote dest IP addr has changed:

- broadcast the updated route entry in RIP responses to other interfaces



Receive:

After receiving a RIP response

- Check whether destIP of this RIP response is srcIP of one interface
  - if yes, ignroe
  - continue other steps 



