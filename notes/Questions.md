# Questions

Some concepts

- protocol, service, API:

  - a service is a program/ an object(like raft) that executes specific tasks in response to events or requests
  - API: a service that interacts with client for performing certain tasks

  ```go
  // important libraries
  import (
      "fmt"
      "log"
      "net/http"
  )
  // Home function
  func Home()(w http.ResponseWriter, r *http.Request){
  // This is what the function will print.
      fmt.Fprintf(w, "Welcome to Educative Home!")
      
  }
  func return_contact()(w http.ResponseWriter, r *http.Request){
  // This is what the function will print.
      fmt.Fprintf(w, "Email: support@educative.io")
      
  }
  // function to handle requests
  func handleReq() {
  // will call Home function by default. 
      http.HandleFunc("/", Home)
      http.HandleFunc("/contact", return_contact)
  
      log.Fatal(http.ListenAndServe(":8200", nil))
  }
  
  func main() {
  //*****************************
  // starting the API
      handleReq()
  }
  ```

  - protocol can define: 
    - struct of request and response that interacts with API
    - data flow
    - rules to transmit data and update state of each node
  - A server provides services to one or more clients, and a server(hardware) is a computer
  - Applications are processes meant to interact with a users. An application can be multiple processes
  - Web services: To be strict, not a service but an application requiring specific protocol

- layer

  - 

- application = process?

  - 

- driver program

  - DRIVDE = command line





1.Link

- Is ip addr of interface transparent to each link
  - 

- link = local port + remote port + UDPConn?
  - 
- 

​	

2.Link Interface

- Is Link Interface a struct? relationship with `interface` in Go
  - 
- Is Link Interface in the IP layer?
  - Link Layer
- Does Link Interace include both local IP address and remote address?
  - 
- link Interface = local IP addr + remote IP addr + Link + status
  - 
- relationship between Link interface and service/API



3.Packet

- struct of packet is predefined packet header (how to call it in Go)+ data

  - 

- all opertaions like check destination IP addr should be done in the network layer. So when we receive a frame in link layer, we extract data(packet) in the frame send it network layer. Then the node can determine that we should drop the packet or send it to next hop.

  - 

- Types of Packets: RIP(200), TEST Protocol (0)

  - 

- RIP format

  uint16 command;
  uint16 num_entries;
  struct {
      uint32 cost;
      uint32 address;
      uint32 mask;
  } entries[num_entries];

  is this RIP format header of packet? -> body

- TEST Protocol send message inside a packet? format?:



4.Frame

- struct of frame = src MAC + dest MAC + Type + Data(packet)
  - 



5.Upper Interface

(1) applicate all the RegisterHandle and we  store the protocol num and function in the table

(2) When we IP layer receives a packet belonging to that node, we call the handler