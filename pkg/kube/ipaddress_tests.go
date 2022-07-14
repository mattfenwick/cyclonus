package kube

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/networking/v1"
	"net"
)

type ipCidrTestCase struct {
	IP       string
	CIDR     string
	IsMember bool
}

func RunIPAddressTests() {
	Describe("IPAddress and CIDRs", func() {
		It("Determines whether an IPv4 address is in a CIDR", func() {
			testCases := []*ipCidrTestCase{
				{
					IP:       "1.2.3.3",
					CIDR:     "1.2.3.0/24",
					IsMember: true,
				},
				{
					IP:       "1.2.3.3",
					CIDR:     "1.2.3.0/28",
					IsMember: true,
				},
				{
					IP:       "1.2.3.3",
					CIDR:     "1.2.3.0/30",
					IsMember: true,
				},
				{
					IP:       "1.2.3.3",
					CIDR:     "1.2.3.0/31",
					IsMember: false,
				},
			}
			for _, c := range testCases {
				log.Infof("looking at %+v", c)
				isInCidr, err := IsIPInCIDR(c.IP, c.CIDR)
				Expect(err).To(BeNil())
				Expect(isInCidr).To(Equal(c.IsMember))
			}
		})

		It("Determines whether an IPv6 address is in a CIDR", func() {
			testCases := []*ipCidrTestCase{
				{
					IP:       "fd00:10:244:a8:96fd:be93:52d8:6b85",
					CIDR:     "fd00:10:244:a8:96fd:be93:52d8:6b00/120",
					IsMember: true,
				},
				{
					IP:       "fd00:10:244:a8:96fd:be93:52d8:6b85",
					CIDR:     "fd00:10:244:a8:96fd:be93:52d8:6b80/124",
					IsMember: true,
				},
				{
					IP:       "fd00:10:244:a8:96fd:be93:52d8:6b90",
					CIDR:     "fd00:10:244:a8:96fd:be93:52d8:6b80/124",
					IsMember: false,
				},
				{
					IP:       "2001:0db8::1.2.3.4",
					CIDR:     "2001:db8::/32",
					IsMember: true,
				},
				{
					IP:       "2001:0db9::",
					CIDR:     "2001:db8::/32",
					IsMember: false,
				},
				{
					IP:       "2001:db8::68",
					CIDR:     "2001:db8::/32",
					IsMember: true,
				},
			}

			for _, c := range testCases {
				log.Infof("looking at %+v", c)
				isInCidr, err := IsIPInCIDR(c.IP, c.CIDR)
				Expect(err).To(BeNil())
				Expect(isInCidr).To(Equal(c.IsMember))
			}
		})

		It("Determines whether an IPv4-mapped IPv6 address id in a CIDR", func() {
			// TODO
			// ::ffff:192.0.2.1
		})

		It("reports an error for malformed IP addresses and CIDRs", func() {
			_, err := IsIPAddressMatchForIPBlock("abc", &v1.IPBlock{
				CIDR:   "1.2.3.4",
				Except: nil,
			})
			Expect(err).ToNot(BeNil())
		})

		It("Handles IPBlocks with no exceptions", func() {
			testCases := []struct {
				IP      string
				IPBlock *v1.IPBlock
				IsMatch bool
			}{
				{
					IP: "1.2.3.3",
					IPBlock: &v1.IPBlock{
						CIDR:   "1.2.3.0/24",
						Except: nil,
					},
					IsMatch: true,
				},
				{
					IP: "1.2.3.3",
					IPBlock: &v1.IPBlock{
						CIDR:   "1.2.3.0/28",
						Except: nil,
					},
					IsMatch: true,
				},
				{
					IP: "1.2.3.3",
					IPBlock: &v1.IPBlock{
						CIDR:   "1.2.3.0/30",
						Except: nil,
					},
					IsMatch: true,
				},
				{
					IP: "1.2.3.3",
					IPBlock: &v1.IPBlock{
						CIDR:   "1.2.3.0/31",
						Except: nil,
					},
					IsMatch: false,
				},
			}
			for _, c := range testCases {
				log.Infof("looking at %+v", c)
				isMatch, err := IsIPAddressMatchForIPBlock(c.IP, c.IPBlock)
				Expect(err).To(BeNil())
				Expect(isMatch).To(Equal(c.IsMatch))
			}
		})

		It("Handles IPBlocks with exceptions", func() {
			testCases := []struct {
				IP      string
				IPBlock *v1.IPBlock
				IsMatch bool
			}{
				{
					IP: "1.2.3.3",
					IPBlock: &v1.IPBlock{
						CIDR:   "1.2.3.0/28",
						Except: []string{"1.2.3.0/30"},
					},
					IsMatch: false,
				},
				{
					IP: "1.2.3.4",
					IPBlock: &v1.IPBlock{
						CIDR:   "1.2.3.0/28",
						Except: []string{"1.2.3.4/30"},
					},
					IsMatch: false,
				},
				{
					IP: "1.2.3.3",
					IPBlock: &v1.IPBlock{
						CIDR:   "1.2.3.0/28",
						Except: []string{"1.2.3.4/30"},
					},
					IsMatch: true,
				},
			}
			for _, c := range testCases {
				log.Infof("looking at %+v", c)

				isMatchWithoutExcept, err := IsIPAddressMatchForIPBlock(c.IP, &v1.IPBlock{
					CIDR:   c.IPBlock.CIDR,
					Except: nil,
				})
				Expect(err).To(BeNil())
				Expect(isMatchWithoutExcept).To(Equal(true))

				isMatchWitExcept, err := IsIPAddressMatchForIPBlock(c.IP, c.IPBlock)
				Expect(err).To(BeNil())
				Expect(isMatchWitExcept).To(Equal(c.IsMatch))
			}
		})
	})

	Describe("Make CIDR from IPAddress", func() {
		It("should build normalized IPV4 CIDRs correctly", func() {
			testCases := []struct {
				IP       string
				Zeroes   int
				Expected string
			}{
				{
					IP:       "255.255.255.255",
					Zeroes:   0,
					Expected: "255.255.255.255/32",
				},
				{
					IP:       "255.255.255.255",
					Zeroes:   1,
					Expected: "255.255.255.254/31",
				},
				{
					IP:       "255.255.255.255",
					Zeroes:   2,
					Expected: "255.255.255.252/30",
				},
				{
					IP:       "255.255.255.255",
					Zeroes:   4,
					Expected: "255.255.255.240/28",
				},
				{
					IP:       "255.255.255.255",
					Zeroes:   8,
					Expected: "255.255.255.0/24",
				},
				{
					IP:       "255.255.255.255",
					Zeroes:   16,
					Expected: "255.255.0.0/16",
				},
			}
			for _, tc := range testCases {
				fmt.Printf("%+v\n", net.ParseIP(tc.IP))
				actual := MakeCIDRFromZeroes(tc.IP, tc.Zeroes)
				Expect(actual).To(Equal(tc.Expected))
			}
		})

		It("should build normalized IPV6 CIDRs correctly", func() {
			testCases := []struct {
				IP       string
				Bits     int
				Expected string
			}{
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     128,
					Expected: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     127,
					Expected: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe/127",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     126,
					Expected: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffc/126",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     124,
					Expected: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:fff0/124",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     120,
					Expected: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ff00/120",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     112,
					Expected: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:0/112",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     96,
					Expected: "ffff:ffff:ffff:ffff:ffff:ffff::/96",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     80,
					Expected: "ffff:ffff:ffff:ffff:ffff::/80",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     64,
					Expected: "ffff:ffff:ffff:ffff::/64",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     48,
					Expected: "ffff:ffff:ffff::/48",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     32,
					Expected: "ffff:ffff::/32",
				},
				{
					IP:       "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
					Bits:     16,
					Expected: "ffff::/16",
				},
			}
			for _, tc := range testCases {
				fmt.Printf("%+v\n", net.ParseIP(tc.IP))
				actual := MakeCIDRFromOnes(tc.IP, tc.Bits)
				Expect(actual).To(Equal(tc.Expected))
			}
		})
	})
}
