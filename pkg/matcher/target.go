package matcher

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Target represents a NetworkPolicySpec.PodSelector, which is in a namespace
type Target struct {
	Namespace   string
	PodSelector metav1.LabelSelector
	Peers       []PeerMatcher
	SourceRules []*networkingv1.NetworkPolicy
	primaryKey  string
}

func (t *Target) String() string {
	return t.GetPrimaryKey()
}

func (t *Target) IsMatch(namespace string, podLabels map[string]string) bool {
	return t.Namespace == namespace && kube.IsLabelsMatchLabelSelector(podLabels, t.PodSelector)
}

// CombinePeerMatchers creates a new Target combining the egress and ingress rules
// of the two original targets.  Neither input is modified.
// The Primary Keys of the two targets must match.
func (t *Target) Combine(other *Target) *Target {
	myPk := t.GetPrimaryKey()
	otherPk := other.GetPrimaryKey()
	if myPk != otherPk {
		panic(errors.Errorf("cannot combine targets: primary keys differ -- '%s' vs '%s'", myPk, otherPk))
	}

	return &Target{
		Namespace:   t.Namespace,
		PodSelector: t.PodSelector,
		Peers:       append(t.Peers, other.Peers...),
		SourceRules: append(t.SourceRules, other.SourceRules...),
	}
}

// The primary key is a deterministic combination of PodSelector and namespace
func (t *Target) GetPrimaryKey() string {
	if t.primaryKey == "" {
		t.primaryKey = fmt.Sprintf(`{"Namespace": "%s", "PodSelector": %s}`, t.Namespace, kube.SerializeLabelSelector(t.PodSelector))
	}
	return t.primaryKey
}

// CombineTargetsIgnoringPrimaryKey creates a new target from the given namespace and pod selector,
// and combines all the edges and source rules from the original targets into the new target.
func CombineTargetsIgnoringPrimaryKey(namespace string, podSelector metav1.LabelSelector, targets []*Target) *Target {
	if len(targets) == 0 {
		return nil
	}
	target := &Target{
		Namespace:   namespace,
		PodSelector: podSelector,
		Peers:       targets[0].Peers,
		SourceRules: targets[0].SourceRules,
	}
	for _, t := range targets[1:] {
		target.Peers = append(target.Peers, t.Peers...)
		target.SourceRules = append(target.SourceRules, t.SourceRules...)
	}
	return target
}
