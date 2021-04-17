package worker

import (
	"fmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"net"
	"strconv"
)

type Batch struct {
	Namespace string
	Pod       string
	Container string
	Requests  []*Request
}

func (b *Batch) Key() string {
	return fmt.Sprintf("%s/%s/%s", b.Namespace, b.Pod, b.Container)
}

func (b *Batch) IsValid() error {
	for _, r := range b.Requests {
		if !protocols[r.Protocol] {
			return errors.Errorf("invalid protocol %+v", r)
		}
	}
	return nil
}

type Result struct {
	Request *Request
	Output  string
	Error   string
}

func (r *Result) IsSuccess() bool {
	return r.Error == ""
}

type Request struct {
	Key      string
	Protocol v1.Protocol
	Host     string
	Port     int
}

func (r *Request) Address() string {
	return net.JoinHostPort(r.Host, strconv.Itoa(r.Port))
}

func (r *Request) Command() []string {
	switch r.Protocol {
	case v1.ProtocolSCTP:
		return []string{"/agnhost", "connect", r.Address(), "--timeout=1s", "--protocol=sctp"}
	case v1.ProtocolTCP:
		return []string{"/agnhost", "connect", r.Address(), "--timeout=1s", "--protocol=tcp"}
	case v1.ProtocolUDP:
		return []string{"/agnhost", "connect", r.Address(), "--timeout=1s", "--protocol=udp"}
	default:
		panic(errors.Errorf("protocol %s not supported", r.Protocol))
	}
}
