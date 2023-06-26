package utils

import "net"

func ReadFromUDPConnection(conn *net.UDPConn, bufferSize int) ([]byte, *net.UDPAddr, error) {
	buffer := make([]byte, bufferSize)
	_, src, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, nil, err
	}
	return buffer, src, nil
}
