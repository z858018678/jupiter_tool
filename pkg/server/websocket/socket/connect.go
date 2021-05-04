package xsocket

func (s *Socket) Run() error {
	go s.send()
	go s.recv()
	return nil
}

func (s *Socket) Stopped() <-chan struct{} {
	return s.opts.ctx.Done()
}

func (s *Socket) Stop() error {
	select {
	case <-s.Stopped():
	default:
		s.opts.cancel()
	}
	return nil
}
