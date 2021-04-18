package kube

import (
	"fmt"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	"net"
)

func IsIPInCIDR(ip string, cidr string) (bool, error) {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, errors.Wrapf(err, "unable to parse CIDR '%s'", cidr)
	}
	trafficIP := net.ParseIP(ip)
	if trafficIP == nil {
		return false, errors.Errorf("unable to parse IP '%s'", ip)
	}
	return cidrNet.Contains(trafficIP), nil
}

func IsIPAddressMatchForIPBlock(ip string, ipBlock *networkingv1.IPBlock) (bool, error) {
	isInCidr, err := IsIPInCIDR(ip, ipBlock.CIDR)
	if err != nil {
		return false, err
	}
	if !isInCidr {
		return false, nil
	}
	for _, except := range ipBlock.Except {
		isInExcept, err := IsIPInCIDR(ip, except)
		if err != nil {
			return false, err
		}
		if isInExcept {
			return false, nil
		}
	}
	return true, nil
}

func IsIPV4Address(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return true
		case ':':
			return false
		}
	}
	panic(errors.Errorf("address %s is neither IPv4 nor IPv6", s))
}

func MakeCIDRFromZeroes(ipString string, zeroes int) string {
	if IsIPV4Address(ipString) {
		return makeCidr(ipString, 32-zeroes, 32)
	}
	return makeCidr(ipString, 128-zeroes, 128)
}

func MakeCIDRFromOnes(ipString string, ones int) string {
	if IsIPV4Address(ipString) {
		return makeCidr(ipString, ones, 32)
	}
	return makeCidr(ipString, ones, 128)
}

func makeCidr(ipString string, ones int, bits int) string {
	mask := net.CIDRMask(ones, bits)
	ip := net.ParseIP(ipString)
	return fmt.Sprintf("%s/%d", ip.Mask(mask).String(), ones)
}
