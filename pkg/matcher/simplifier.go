package matcher

import (
	"github.com/pkg/errors"
	"sort"
)

type Simplifier struct {
	MatchesAll              bool
	PortsForAllPeersMatcher *PortsForAllPeersMatcher
	IPs                     []*IPPeerMatcher
	Pods                    []*PodPeerMatcher
}

func NewSimplifier(matchers []PeerMatcher) *Simplifier {
	s := &Simplifier{}
	var alls []*PortsForAllPeersMatcher
	for _, matcher := range matchers {
		switch a := matcher.(type) {
		case *AllPeersMatcher:
			s.MatchesAll = true
		case *PortsForAllPeersMatcher:
			alls = append(alls, a)
		case *NonePeerMatcher:
			// nothing to do
		case *IPPeerMatcher:
			s.IPs = append(s.IPs, a)
		case *PodPeerMatcher:
			s.Pods = append(s.Pods, a)
		default:
			panic(errors.Errorf("invalid matcher type %T", matcher))
		}
	}
	s.PortsForAllPeersMatcher = simplifyAlls(alls)
	s.IPs = simplifyIPMatchers(s.IPs)
	s.Pods = simplifyPodMatchers(s.Pods)
	if s.PortsForAllPeersMatcher != nil {
		s.IPs, s.Pods = simplifyIPsAndPodsIntoAlls(s.PortsForAllPeersMatcher, s.IPs, s.Pods)
	}
	return s
}

func simplifyAlls(alls []*PortsForAllPeersMatcher) *PortsForAllPeersMatcher {
	if len(alls) == 0 {
		return nil
	}
	port := alls[0].Port
	for _, a := range alls[1:] {
		port = CombinePortMatchers(port, a.Port)
	}
	return &PortsForAllPeersMatcher{Port: port}
}

func simplifyPodMatchers(pms []*PodPeerMatcher) []*PodPeerMatcher {
	grouped := map[string]*PodPeerMatcher{}
	for _, pm := range pms {
		key := pm.PrimaryKey()
		if _, ok := grouped[key]; !ok {
			grouped[key] = pm
		} else {
			grouped[key] = CombinePodPeerMatcher(grouped[key], pm)
		}
	}
	var simplified []*PodPeerMatcher
	for _, pm := range grouped {
		simplified = append(simplified, pm)
	}
	sort.Slice(simplified, func(i, j int) bool {
		return simplified[i].PrimaryKey() < simplified[j].PrimaryKey()
	})
	return simplified
}

func simplifyIPMatchers(ims []*IPPeerMatcher) []*IPPeerMatcher {
	grouped := map[string]*IPPeerMatcher{}
	for _, im := range ims {
		key := im.PrimaryKey()
		if _, ok := grouped[key]; !ok {
			grouped[key] = im
		} else {
			grouped[key] = CombineIPPeerMatcher(grouped[key], im)
		}
	}
	var simplified []*IPPeerMatcher
	for _, pm := range grouped {
		simplified = append(simplified, pm)
	}
	sort.Slice(simplified, func(i, j int) bool {
		return simplified[i].PrimaryKey() < simplified[j].PrimaryKey()
	})
	return simplified
}

func simplifyIPsAndPodsIntoAlls(all *PortsForAllPeersMatcher, ips []*IPPeerMatcher, pods []*PodPeerMatcher) ([]*IPPeerMatcher, []*PodPeerMatcher) {
	var newIps []*IPPeerMatcher
	for _, ip := range ips {
		isEmpty, remainingPorts := SubtractPortMatchers(ip.Port, all.Port)
		if isEmpty {
			// nothing to do
		} else {
			newIps = append(newIps, &IPPeerMatcher{
				IPBlock: ip.IPBlock,
				Port:    remainingPorts,
			})
		}
	}
	var newPods []*PodPeerMatcher
	for _, pod := range pods {
		isEmpty, remainingPorts := SubtractPortMatchers(pod.Port, all.Port)
		if isEmpty {
			// nothing to do
		} else {
			newPods = append(newPods, &PodPeerMatcher{
				Namespace: pod.Namespace,
				Pod:       pod.Pod,
				Port:      remainingPorts,
			})
		}
	}
	return newIps, newPods
}

func (s *Simplifier) SimplifiedMatchers() []PeerMatcher {
	if s.MatchesAll {
		return []PeerMatcher{AllPeersPorts}
	}
	if s.PortsForAllPeersMatcher == nil && len(s.IPs) == 0 && len(s.Pods) == 0 {
		return []PeerMatcher{NoPeers}
	}
	var matchers []PeerMatcher
	if s.PortsForAllPeersMatcher != nil {
		matchers = append(matchers, s.PortsForAllPeersMatcher)
	}
	for _, ip := range s.IPs {
		matchers = append(matchers, ip)
	}
	for _, pod := range s.Pods {
		matchers = append(matchers, pod)
	}
	return matchers
}

func CombinePortMatchers(a PortMatcher, b PortMatcher) PortMatcher {
	switch l := a.(type) {
	case *AllPortMatcher:
		return a
	case *SpecificPortMatcher:
		switch r := b.(type) {
		case *AllPortMatcher:
			return b
		case *SpecificPortMatcher:
			return l.Combine(r)
		default:
			panic(errors.Errorf("invalid Port type %T", b))
		}
	default:
		panic(errors.Errorf("invalid Port type %T", a))
	}
}

func SubtractPortMatchers(a PortMatcher, b PortMatcher) (bool, PortMatcher) {
	switch l := a.(type) {
	case *AllPortMatcher:
		return true, nil
	case *SpecificPortMatcher:
		switch r := b.(type) {
		case *AllPortMatcher:
			return false, b
		case *SpecificPortMatcher:
			return false, l.Subtract(r)
		default:
			panic(errors.Errorf("invalid Port type %T", b))
		}
	default:
		panic(errors.Errorf("invalid Port type %T", a))
	}
}

func CombinePodPeerMatcher(a *PodPeerMatcher, b *PodPeerMatcher) *PodPeerMatcher {
	if a.PrimaryKey() != b.PrimaryKey() {
		panic(errors.Errorf("cannot combine PodPeerMatchers of different pks: %s vs. %s", a.PrimaryKey(), b.PrimaryKey()))
	}
	return &PodPeerMatcher{
		Namespace: a.Namespace,
		Pod:       a.Pod,
		Port:      CombinePortMatchers(a.Port, b.Port),
	}
}

func CombineIPPeerMatcher(a *IPPeerMatcher, b *IPPeerMatcher) *IPPeerMatcher {
	if a.PrimaryKey() != b.PrimaryKey() {
		panic(errors.Errorf("unable to combine IPPeerMatcher values with different primary keys: %s vs %s", a.PrimaryKey(), b.PrimaryKey()))
	}
	return &IPPeerMatcher{
		IPBlock: a.IPBlock,
		Port:    CombinePortMatchers(a.Port, b.Port),
	}
}
