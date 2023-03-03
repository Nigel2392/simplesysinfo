package simplesysinfo

import (
	"net"
	"strings"
)

const nullAddr = "00:00:00:00:00:00"

func GetMACAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nullAddr, err
	}
	var currentIP, currentNetworkHardwareName string
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		// = GET LOCAL IP ADDRESS
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				currentIP = ipnet.IP.String()
			}
		}
	}
	// get all the system's or local machine's network interfaces
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {
		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				// only interested in the name with current IP address
				if strings.Contains(addr.String(), currentIP) {
					currentNetworkHardwareName = interf.Name
				}
			}
		}
	}
	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)
	if err != nil {
		return nullAddr, err
	}
	macAddress := netInterface.HardwareAddr
	// verify if the MAC address can be parsed properly
	hwAddr, err := net.ParseMAC(macAddress.String())
	if err != nil {
		return nullAddr, err
	}
	return hwAddr.String(), nil
}
