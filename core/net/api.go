package net

func StopTcpClients() {
	logger.Infof("<<<clients is stop start>>>")
	clients.Range(func(k, v interface{}) bool {
		if s, ok := v.(*TCPClient); ok {
			s.Close()
		}
		return true
	})
	logger.Infof("<<<clients is stop over>>>")
}
