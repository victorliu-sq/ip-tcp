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



kill process

```shell
sudo install lsof
#(get process who's listening on 17001)
sudo lsof -i -P -n | grep 17001 

kill -9 pid
```





# Three way handshake

## Part1

server()

```shell
./node ./nets/routeAggregation/tree/B.lnx

a 80
```



client() - reference node

```shell
./tools/ref_tcp_node ./nets/routeAggregation/tree/A.lnx

c 10.0.0.9 80

c 10.0.0.14 80
```



client() - my node

```shell
./node ./nets/routeAggregation/tree/A.lnx

c 10.0.0.9 80

c 10.0.0.14 80
```



## Part2

server()

```shell
./node ./nets/routeAggregation/tree/B.lnx

c 10.0.0.1 80

s 1 dsjakldjsla
```



client()

```shell
./tcp_tools/ref_tcp_node ./nets/routeAggregation/tree/A.lnx

a 80

r 1 1000
```



# Retransmit

ref_tcp_node +A 

```shell
./tools/ref_tcp_node ./nets/routeAggregation/tree/A.lnx

./node ./nets/routeAggregation/tree/A.lnx

a 80
```



lossy_ip_node + B

```shell
./tools/lossy_ip_node ./nets/routeAggregation/tree/B.lnx

lossy 0.4
```



tcp_node + C

```shell
./node ./nets/routeAggregation/tree/C.lnx

c 10.0.0.1 80
```



## Full Window

lossy_ip_node + B

```shell
./tools/lossy_ip_node ./nets/routeAggregation/tree/B.lnx

lossy 0.4
```



ref_tcp_node +A 

```shell
./node ./nets/routeAggregation/tree/A.lnx

a 80

r 1 5 n

r 1 35 y

# receive 105 bytes
r 1 105 y

# receive 1400 bytes
r 1 1400 y
# result should be 12345 6789a bcdef ghijk lmnop qrstu vwxyz
```



tcp_node + C

```shell
./node ./nets/routeAggregation/tree/C.lnx
./tools/ref_tcp_node ./nets/routeAggregation/tree/C.lnx

c 10.0.0.1 80

s 0 12345

s 0 123456

s 0 0123456789

# send 35 bytes
s 0 123456789abcdefghijklmnopqrstuvwxyz

# send 105 bytes
s 0 123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz

# send 1400 bytes
s 0 123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz
```



# Bidirection

A

```shell
# send 105 bytes
s 1 123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz
```



C

```shell
# receive 105 bytes
r 0 105 y
```



Fixed_Bug: 

//first byte acked doesn't mean all bytes in payload have been acked

// we must add it to map as many as possible (length of payload can get out of range)

00:21:47 debugger.go:25: [Server] Current recv buffer content: *op**
00:21:47 debugger.go:25: [Server] 10.0.0.1:80 sent to 10.0.0.10:1024, SEQ: 1596996164, ACK: 2596996185, Win: 5
00:21:47 debugger.go:25: [Server] 10.0.0.1:80 receive from 10.0.0.10:1024, SEQ: 2596996188, ACK 1596996164, Payload qrs
old una 2596996185
new una 2596996185
00:21:47 debugger.go:25: [Server] Current recv buffer content: sopqr
00:21:47 debugger.go:25: [Server] 10.0.0.1:80 sent to 10.0.0.10:1024, SEQ: 1596996164, ACK: 2596996185, Win: 5



# SR Test

## Node-A

```shell
./node ./nets/routeAggregation/tree/A.lnx
./tools/ref_tcp_node ./nets/routeAggregation/tree/A.lnx

a 80

s 1 12345

s 1 0123456789

# send 35 bytes
s 1 123456789abcdefghijklmnopqrstuvwxyz

# send 105 bytes
s 1 123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz

# send 1400 bytes
s 1 123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz


# send file
sf 1400file.txt 10.0.1.1 90

sf sendfile 10.0.1.1 90

# reference node C sends file to A NOTICE we need to lower the lossy rate !
sf sendfile 10.0.0.1 90

# shutdown write
sd 0 write
sd 0 read
sd 0 both
```



## Node-B

```shell
./tools/lossy_ip_node ./nets/routeAggregation/tree/B.lnx

lossy 0.4
```



## Node-C

```shell
./node ./nets/routeAggregation/tree/C.lnx
./tools/ref_tcp_node ./nets/routeAggregation/tree/C.lnx

c 10.0.0.1 80

r 0 5 n

r 0 35 y

# receive 105 bytes
r 0 105 y

# receive 1400 bytes
r 0 1400 y

# receive file of 1400 bytes
rf 1400fileR 90
# receive file of 1MB
rf rcvfile 90
# receive file and print its content on the screen
rf /dev/stdout 90


# shutdown write
sd 0 read
sd 1 read
```



# Close

```shell
# A
./node ./nets/routeAggregation/tree/A.lnx

a 80

# C
./node ./nets/routeAggregation/tree/C.lnx

c 10.0.0.1 80

# A
cl 1

# ------------------------------------------------------------------------
# Test for bidirectional
# A
r 1 105 y

# C 
s 0 123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz123456789abcdefghijklmnopqrstuvwxyz

# ------------------------------------------------------------------------

# C
cl 0


```





# SendFile, HashNum

My Example

```shell
rm 1400fileR

# A
./node ./nets/routeAggregation/tree/A.lnx
rf 1400fileR 90

# C
./node ./nets/routeAggregation/tree/C.lnx
sf 1400file 10.0.0.1 90

# check bytes of 2 files
ls -la 1400file
ls -la 1400fileR

# check hash of 2 files
sha1sum 1400file 1400fileR
```





Other Example

```shell
# create a file of 1MB
dd if=/dev/urandom bs=1M count=1 | base64 -w 0 > sendfile
# create a file of 100MB
dd if=/dev/urandom bs=100M count=1 | base64 -w 0 > sendfile

rm rcvfile
rm 1400fileR

# A
./node ./nets/routeAggregation/tree/A.lnx
rf rcvfile 90
rf 1400fileR 90

# C
./node ./nets/routeAggregation/tree/C.lnx
sf sendfile 10.0.0.1 90
sf 1400file 10.0.0.1 90

# check bytes of 2 files
ls -la sendfile 
ls -la rcvfile

# check hash of 2 files
sha1sum 1400file.txt 1400fileR
sha1sum sendfile rcvfile
```

