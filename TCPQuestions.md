# Sliding Window Protocol Comment

**Be sure that you can accept out-of-order packets.** That is, a packet’s sequence number doesn’t have to be exactly the sequence number of the start of the window. 

It can be fully contained within the window, somewhere in the middle. The easiest way to handle such packets is to place them on a queue of potentially valid packets, and then deal with them once the window has caught up to the beginning of that segment’s sequence number.

You are not required to implement slow start, but you should detect dropped or un’acked packets and adjust your flow accordingly.

You should strictly adhere to the flow control window as specified in the RFC, e.g. do not send packets outside of your window, etc. Similarly, you should implement zero window probing to ensure your sender can recover when the receiver’s window is full. Overall, your goal is to ensure reliability—all data must get to its destination in order, uncorrupted.

As you implement your protocol, keep in mind how sliding windows will interact with the rest of TCP. For example, a call to CLOSE only closes data flow in one direction. Because data will still be flowing in the other direction, the closed side will need to send acknowledgments and window updates until both sides have closed.



# Buffer

Flow control:What is the smallest unit for buffer? Byte or segment? -> bytes

# Sender:

-Is the window size of sender set by user? -> Set by the win_size in ACK msg from receiver

-Does each segment in the window identified by the seq#? ->Yes

-Does the win_size of conn in sender need to be changed whenever it receives an ACK from receiver?  -> Yes

-Does the sender need to keep a data structure for to accumulate segments to send ACK -> Yes, retransmission queue

-Retransmission Queue: element: seq#, []byte, timestamp

-What will happen if the buffer cannot store all bytes to load -> return # of bytes written by the client.

## SENDER_Algorithm:

UNA NXT LBW

Window: [UNA NXT - 1]

- conn.Write loads bytes into buffer, tries to send out bytes w/in the window, moves NXT
- Sender receives ACK w/in the window, tries to moves forward UNA as much as possible, tries to send out segments(add them to retransmission queue) …
- Zero-probe, if the win_size == 0, send out 1 segment
  -Should UNA and NXT be moved together when sender receives ACK? -> No



# Receiver:

Does the receiver need to keep a data structure for to accumulate segments to send ACK -> No, loss of ACK will make the sender to retransmit a packet to get another ACK

## RECEIVER_Algorithm:

LBR NXTWindow: [NXT, NXT + win_size - 1]

win_size = 2 ^ 16 - ((NXT - 1) - LBR)

- Receiver receives segments w/in the window, sends back ACK and tries to move NXT as much as possible
- conn.Read() removes bytes from window (upper bound is given by all the bytes in the buffer)