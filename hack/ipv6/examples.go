package main

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"net"
)

func main() {
	ips := []string{
		"1.2.3.4",
		"fd00:10:244:a8:96fd:be93:52d8:6b85",
		"fc00:f853:ccd:e793::2",
		"fc00:f853:ccd:e793::4",
		"fc00:f853:ccd:e793::9999",
		"fc00:f853:ccd:e793::aaaa:9999",
		"fc00:f853:ccd:e793::aaaa:bbbb:9999",
		"fc00::",
		"ffff:ffff:ffff:ffff::",
		"2001:db8::1",
		"1:2:3:4:0:0:0:0",
		"::",
		"::12:34",
		"2001:db9::",
	}
	for _, i := range ips {
		ip := net.ParseIP(i)
		if ip == nil {
			panic(errors.Errorf("unable to parse ip address %s", i))
		}
		fmt.Printf("%s\n - parsed: %s\n - bytes: %+v\n", i, ip.String(), []byte(ip))
	}

	fmt.Println()

	invalidIPs := []string{
		"fc00:f853:ccd:e793::99999999",
		"fc00:f853:ccd:e793::aaaa:bbbb:cccc:9999",
		"2001:db9",
		"2001:0db9",
	}
	for _, i := range invalidIPs {
		ip := net.ParseIP(i)
		if ip != nil {
			panic(errors.Errorf("expected ip address %s to be invalid", i))
		}
		fmt.Printf("ip address %s invalid as expected\n", i)
	}

	fmt.Println()

	cidrs := []string{
		"fc00::/7",
		"fc00::/48",
		"ffff:ffff:ffff:ffff::/64",
		"ffff:ffff:ffff:ffff:0:0::/64",
		"ffff:ffff:ffff:ffff:0:0:0:0/64",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128",
	}
	for _, i := range cidrs {
		ip, ipnet, err := net.ParseCIDR(i)
		utils.DoOrDie(err)
		fmt.Printf("%s\n - ip: %s\n - bytes: %+v\n - net bytes: %+v\n - net mask: %+v\n",
			i, ip.String(), []byte(ip), []byte(ipnet.IP), []byte(ipnet.Mask))
	}

	fmt.Println()

	invalidCidrs := []string{
		"ffff:ffff:ffff:ffff:0:0:0:0::/64",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff::/128",
	}
	for _, i := range invalidCidrs {
		_, _, err := net.ParseCIDR(i)
		if err == nil {
			panic(errors.Errorf("expected cidr %s to be invalid", i))
		}
		fmt.Printf("cidr %s invalid as expected: %+v\n", i, err)
	}
}
