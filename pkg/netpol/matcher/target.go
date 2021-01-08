package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

// Target represents a NetworkPolicySpec.PodSelector, which is in a namespace
type Target struct {
	Namespace   string
	PodSelector metav1.LabelSelector
	Edge        EdgeMatcher
	SourceRules []*networkingv1.NetworkPolicy
	primaryKey  string
}

func (t *Target) String() string {
	return t.GetPrimaryKey()
}

func (t *Target) IsMatch(namespace string, podLabels map[string]string) bool {
	return t.Namespace == namespace && kube.IsLabelsMatchLabelSelector(podLabels, t.PodSelector)
}

// CombineEdgeMatchers creates a new Target combining the egress and ingress rules
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
		Edge:        CombineEdgeMatchers(t.Edge, other.Edge),
		SourceRules: append(t.SourceRules, other.SourceRules...),
	}
}

// The primary key is a deterministic combination of PodSelector and namespace
func (t *Target) GetPrimaryKey() string {
	if t.primaryKey == "" {
		t.primaryKey = fmt.Sprintf(`{"Namespace": "%s", "PodSelector": %s}`, t.Namespace, SerializeLabelSelector(t.PodSelector))
	}
	return t.primaryKey
}

// SerializeLabelSelector deterministically converts a metav1.LabelSelector
// into a string
func SerializeLabelSelector(ls metav1.LabelSelector) string {
	var labelKeys []string
	for key := range ls.MatchLabels {
		labelKeys = append(labelKeys, key)
	}
	sort.Slice(labelKeys, func(i, j int) bool {
		return labelKeys[i] < labelKeys[j]
	})
	var keyVals []string
	for _, key := range labelKeys {
		keyVals = append(keyVals, fmt.Sprintf("%s: %s", key, ls.MatchLabels[key]))
	}
	// this is weird, but use an array to make the order deterministic
	bytes, err := json.Marshal([]interface{}{"MatchLabels", keyVals, "MatchExpression", ls.MatchExpressions})
	if err != nil {
		panic(errors.Wrapf(err, "unable to marshal json"))
	}
	return string(bytes)
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
		Edge:        targets[0].Edge,
		SourceRules: targets[0].SourceRules,
	}
	for _, t := range targets[1:] {
		target.Edge = CombineEdgeMatchers(target.Edge, t.Edge)
		target.SourceRules = append(target.SourceRules, t.SourceRules...)
	}
	return target
}
