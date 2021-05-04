package xsocket

func (s *Socket) heartbeat(data []byte) {
	if string(data) != string(s.opts.heartbeatMsg) {
		return
	}

	s.opts.hbTicker.Reset(s.opts.heartbeatInterval)
}

func (s *Socket) checkHeartbeat() {
	<-s.opts.hbTicker.C
	s.Stop()
}
