package main

import (
	"fmt"
	"net"
)

func getMacAddr(ifname string) ([]byte, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, ifa := range ifas {
		if ifa.Name != ifname {
			continue
		}
		fmt.Printf("MAC: %s\n", ifa.HardwareAddr.String())
		return ifa.HardwareAddr, nil
	}
	return nil, nil
}

func macAddressToChaddr(addr []byte) uint16 {
	var n uint16
	for i := 0; i < 6; i += 1 {
		n = ((n << 8) | uint16(addr[5-i]))
	}
	return n
}
