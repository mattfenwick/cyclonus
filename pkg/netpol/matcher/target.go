package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

// Combine creates a new Target combining the egress and ingress rules
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
		Edge:        Combine(t.Edge, other.Edge),
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
		log.Fatalf("%+v", err)
	}
	return string(bytes)
}
