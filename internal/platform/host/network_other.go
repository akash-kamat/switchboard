//go:build windows || darwin

package host

import (
	"net"
	"strings"
)

func localIP() string {
	connection, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 80})
	if err == nil {
		defer connection.Close()
		if address, ok := connection.LocalAddr().(*net.UDPAddr); ok && usableIPv4(address.IP) {
			return address.IP.String()
		}
	}
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		name := strings.ToLower(iface.Name)
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 || strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "veth") {
			continue
		}
		addresses, _ := iface.Addrs()
		for _, address := range addresses {
			ip, _, parseErr := net.ParseCIDR(address.String())
			if parseErr == nil && usableIPv4(ip) {
				return ip.String()
			}
		}
	}
	return "—"
}

func usableIPv4(ip net.IP) bool {
	return ip.To4() != nil && !ip.IsLoopback() && !ip.IsUnspecified() && !ip.IsLinkLocalUnicast()
}
