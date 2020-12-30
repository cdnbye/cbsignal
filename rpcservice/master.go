package rpcservice

type Master struct {
	addr string         // ip:port

}

func (s *Master)Init(addr string) {
	s.addr = addr
}

func (s *Master)IsMaster() bool {
	return true
}

func (s *Master)Addr() string {
	return s.addr
}

// 想master发送心跳包，携带addr，在响应中获取peers的addr
func (s *Master)Heartbeat() {

}
