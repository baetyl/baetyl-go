package utils

import (
	"fmt"
	"net"
)

// GetAvailablePort finds an available port
func GetAvailablePort(host string) (int, error) {
	address, err := net.ResolveTCPAddr("tcp", host+":0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func CheckPortAvailable(ip string, port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return false
	}
	l.Close()
	return true
}
