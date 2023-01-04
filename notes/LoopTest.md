# Loop Test

## Build Network

net 2 inx

```shell
./tools/net2lnx nets/loop.net
```



### build 6 ref nodes

```shell
./tools/ref_node src.lnx

./tools/ref_node srcR.lnx

#short can be open in another window
./tools/ref_node short.lnx
# ----------------------------------

./tools/ref_node long1.lnx

./tools/ref_node long2.lnx

./tools/ref_node dstR.lnx

./tools/ref_node dst.lnx
```



### build 6 self-designed node

```shell
./node src.lnx

./node srcR.lnx

./node long1.lnx

#short can be open in another window
./node short.lnx
# ----------------------------------

./node long2.lnx

./node dstR.lnx

./node dst.lnx
```



### build net with nodes of integrated types

```shell
./node src.lnx

./node srcR.lnx

./node long1.lnx

#short can be open in another window
./node short.lnx
# ----------------------------------

./tools/ref_node long2.lnx

./tools/ref_node dstR.lnx

./tools/ref_node dst.lnx
```







```shell
#lr of short
192.168.0.6     192.168.0.6    1
192.168.0.12     192.168.0.6    1
192.168.0.9     192.168.0.3    2
192.168.0.2     192.168.0.3    1
192.168.0.5     192.168.0.5    0
192.168.0.1     192.168.0.3    2
192.168.0.14     192.168.0.6    2
192.168.0.7     192.168.0.3    1
192.168.0.3     192.168.0.3    1
192.168.0.8     192.168.0.3    2
192.168.0.13     192.168.0.6    1
192.168.0.4     192.168.0.4    0

#lr of src
# [dst].14 => 4
# [dstR].6 & .12 &.13 => 3
# [long2].10 & .11 => 2
# [src, long1].1 & .8 & .9 => 1
# [srcR].2 & .3 & .7 => 0
```



## Down

short down 4

```shell
# src can still reach short 
# src -> srcR -> long1 -> long2 -> dstR -> short

# [dst, short].14 .5 => 5
# [dstR].6 & .12 & .13 => 4
# [long2].10 & 11 => 3
# [long1].8 & .9 => 2
# [srcR].2 & .3 & .7 => 1
# [src].1 => 0
```



## Up

```shell
# ask the short node to up its link 0
up 0
```



## Quit the short node

Afterwards, we check the routing tables 

```shell
# lr of src
# [dst].14 => 5
# [dstR].6 & .12 & .13 => 4
# [long2].10 & 11 => 3
# [long1].8 & .9 => 2
# [srcR].2 & .3 & .7 => 1
# [src].1 => 0
```



```shell
# lr of srcR
# [dst].14 => 4
# [dstR].6 & .12 &.13 => 3
# [long2].10 & .11 => 2
# [src, long1].1 & .8 & .9 => 1
# [srcR].2 & .3 & .7 => 0
```



```shell
# lr of long1
# [dst].14 => 3
# [src, dstR].1 & .6 & & .12 & .13 => 2
# [srcR, long2].2 & .3 & .7 & .10 & .11 => 1
# [long1].8 & .9 => 0
```



```shell
# lr of long2
# [src].1 => 3
# [srcR, dst].2 & .3 & .7 & .14 => 2
# [long1, dstR].6 & .8 & .9  & .12 & .13 => 1
# [long2].10 & .11 => 0
```



```shell
# lr of dstR
# [src].1 => 4
# [srcR].2 & .3 & .7 => 3
# [long1].8 & .9 => 2
# [long2, dst].10 & .11 & .14 => 1
# [dstR].6 & .12 & .13 => 0
```



```shell
# lr of dst
# [src].1 => 5
# [srcR].2 & .3 & .7 => 4
# [long1].8 & .9 => 3
# [long2].10 & .11 => 2
# [dstR].6 & .12 & .13 => 1
# [dst].14 => 0
```



## restart short

src sends a packet to dst

```shell
#src send a packet to dst
send 192.168.0.14 0 Hello from src

#src send a packet to .5 port of short
send 192.168.0.5 0 Hello from src

#src send a packet to dst
send 192.168.0.1 0 Hello from dst
```

