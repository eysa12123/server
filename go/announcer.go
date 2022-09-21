package main

import (
	"fmt"
	"net"

	"github.com/brutella/dnssd"
)

func getNetInterfaces() []net.Interface {
	if interfaceList, interfaceListErr := net.Interfaces(); interfaceListErr == nil {
		return interfaceList
	}
	return []net.Interface{}
}

func getInterfaceIPs(iface net.Interface) []net.IP {
	if addrs, err := iface.Addrs(); err == nil {
		result := make([]net.IP, 0)
		for _, addr := range addrs {
			if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
				result = append(result, ip)
			}
		}
		return result
	}
	return nil
}

func configureAnnouncer(ips []net.IP, hostName string, port int) (dnssd.Responder, dnssd.ServiceHandle, error) {
	// We only want to advertise on the interfaces that go with the addresses!
	interfaces := make([]string, 0)

	// Let's check against each of the interfaces. If (one of) the address(es) on that interface matches
	// one that we are announcing, then we want to send it out that address!
InterfacesLoop:
	for _, iface := range getNetInterfaces() {
		// Go through all the IPs associated with each interface.
		for _, ifaceIP := range getInterfaceIPs(iface) {
			// Go through all the IPs that we are broadcasting on.
			for _, ip := range ips {
				// If there is an intersection, then add the interface to the list we will advertise on
				// and then continue!

				if ifaceIP.Equal(ip) {
					interfaces = append(interfaces, iface.Name)
					continue InterfacesLoop
				}
			}
		}
	}
	cfg := dnssd.Config{
		Name:   fmt.Sprintf("RPM Test Server(%d)", port),
		Type:   "_nq._tcp",
		Domain: "local",
		Host:   hostName,
		IPs:    ips,
		Ifaces: interfaces,
		Port:   port,
	}

	dnsService, err := dnssd.NewService(cfg)
	if err != nil {
		return nil, nil, err
	}
	dnsResponder, err := dnssd.NewResponder()
	if err != nil {
		return nil, nil, err
	}
	dnsServiceHandle, err := dnsResponder.Add(dnsService)
	if err != nil {
		return nil, nil, err
	}

	return dnsResponder, dnsServiceHandle, nil
}
