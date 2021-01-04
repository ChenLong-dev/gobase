package utils

import (
	"errors"
	"github.com/ChenLong-dev/gobase/mlog"
	"net"
	"strconv"
	"strings"
)

func GetLocalIPv4() (string, error) {
	var ips []string
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}
	if 0 == len(ips) {
		return "", errors.New("No available ip ")
	}
	return ips[0], nil
}

func GetAddress(host string, port int) string {
	address := strings.Join([]string{host, ":", strconv.Itoa(port)}, "")
	mlog.Infof("[SERVER] address: %v", address)
	return address
}

func GetUnusedLis(host string, port int) (net.Listener, string) {
	address := GetAddress(host, port)
	lis, err := net.Listen("tcp", address)
	for err != nil {
		mlog.Errorf("[SERVER]failed to listen:%v", err)
		port += 1
		address = GetAddress(host, port)
		lis, err = net.Listen("tcp", address)
	}
	return lis, address
}
