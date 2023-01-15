package transport

// Write bytes into SND Buffer
func (conn *VTCPConn) WriteToSNDLoop(content []byte) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	total := uint32(len(content))
	for total > 0 {
		if !conn.snd.IsFull() {
			bnum := conn.snd.WriteIntoBuffer(content)
			total -= bnum
			content = content[bnum:]
			conn.scv.Signal()
			// Print current snd
			conn.snd.PrintSND()
		} else {
			conn.wcv.Wait()
		}
	}
}
