package synthetic

import (
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Request struct {
	Protocol  v1.Protocol
	Port      intstr.IntOrString
	Policies  *matcher.Policy
	Resources *Resources
}

type Result struct {
	Request  *Request
	Ingress  *utils.TruthTable
	Egress   *utils.TruthTable
	Combined *utils.TruthTable
}
