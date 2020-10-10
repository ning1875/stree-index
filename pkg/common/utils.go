package common

import (
	"net"
	"log"
)

var (
	NodeLevelMap = map[string]uint64{
		"a": 4,
		"p": 3,
		"g": 2,
		"o": 1,
	}
	NodeLevelReverseMap = map[uint64]string{
		4: "a",
		3: "p",
		2: "g",
		1: "o",
	}

	//NodeLevelMap = map[string]int{
	//	"a": 4,
	//	"p": 3,
	//	"g": 2,
	//	"o": 1,
	//}
)

func GetLocalIp() string {
	localIp := "empty"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println("GetLocalIp_net.InterfaceAddrs", err)
		return localIp
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				addr := ipnet.IP.String()
				localIp = addr
				return localIp
			}
		}
	}

	return localIp
}
