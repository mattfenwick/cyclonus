package kube

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/networking/v1"
)

func RunIPAddressTests() {
	Describe("IPAddress and CIDRs", func() {
		It("Determines whether an IP address is in a CIDR", func() {
			testCases := []struct {
				IP       string
				CIDR     string
				IsMember bool
			}{
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

		It("Handles IPv6 addresses and CIDRs", func() {
			// 2001:db8::/32
			// 2001:db8::68
			// IPv4-mapped IPv6 ("::ffff:192.0.2.1")
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
}
