package synthetic

import (
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type Request struct {
	Protocol  v1.Protocol
	Port      int
	Policies  *matcher.Policy
	Resources *Resources
}

type Result struct {
	Request  *Request
	Ingress  *utils.TruthTable
	Egress   *utils.TruthTable
	Combined *utils.TruthTable
}
