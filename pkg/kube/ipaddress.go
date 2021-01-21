package kube

import (
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
