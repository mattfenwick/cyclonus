package kube

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
)

// IsNameMatch follows the kube pattern of "empty string means matches All"
// It will return:
//   if matcher is empty: true
//   if objectName and matcher are the same: true
//   otherwise false
func IsNameMatch(objectName string, matcher string) bool {
	if matcher == "" {
		return true
	}
	return objectName == matcher
}

func IsMatchExpressionMatchForLabels(labels map[string]string, exp metav1.LabelSelectorRequirement) bool {
	switch exp.Operator {
	case metav1.LabelSelectorOpIn:
		val, ok := labels[exp.Key]
		if !ok {
			return false
		}
		for _, v := range exp.Values {
			if v == val {
				return true
			}
		}
		return false
	case metav1.LabelSelectorOpNotIn:
		val, ok := labels[exp.Key]
		if !ok {
			// see https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements
			//   even for NotIn -- if the key isn't there, it's not a match
			return false
		}
		for _, v := range exp.Values {
			if v == val {
				return false
			}
		}
		return true
	case metav1.LabelSelectorOpExists:
		_, ok := labels[exp.Key]
		return ok
	case metav1.LabelSelectorOpDoesNotExist:
		_, ok := labels[exp.Key]
		return !ok
	default:
		panic("invalid operator")
	}
}

// IsLabelsMatchLabelSelector matches labels to a kube LabelSelector.
// From the docs:
// > A label selector is a label query over a set of resources. The result of matchLabels and
// > matchExpressions are ANDed. An empty label selector matches all objects. A null
// > label selector matches no objects.
func IsLabelsMatchLabelSelector(labels map[string]string, labelSelector metav1.LabelSelector) bool {
	// From the docs: "The requirements are ANDed."
	//   Therefore, all MatchLabels must be matched.
	for key, val := range labelSelector.MatchLabels {
		if labels[key] != val {
			return false
		}
	}

	// From the docs: "The requirements are ANDed."
	//   Therefore, all MatchExpressions must be matched.
	for _, exp := range labelSelector.MatchExpressions {
		isMatch := IsMatchExpressionMatchForLabels(labels, exp)
		if !isMatch {
			return false
		}
	}

	// From the docs: "An empty label selector matches all objects."
	return true
}

func IsLabelSelectorEmpty(l metav1.LabelSelector) bool {
	return len(l.MatchLabels) == 0 && len(l.MatchExpressions) == 0
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

func LabelSelectorTableLines(selector metav1.LabelSelector) string {
	// TODO doesn't this return the wrong thing when the selector is for a namespace?
	if IsLabelSelectorEmpty(selector) {
		return "all pods"
	}
	var lines []string
	if len(selector.MatchLabels) > 0 {
		lines = append(lines, "Match labels:")
		for key, val := range selector.MatchLabels {
			lines = append(lines, fmt.Sprintf("  %s: %s", key, val))
		}
	}
	if len(selector.MatchExpressions) > 0 {
		lines = append(lines, "Match expressions:")
		for _, exp := range selector.MatchExpressions {
			lines = append(lines, fmt.Sprintf("  %s %s %+v", exp.Key, exp.Operator, exp.Values))
		}
	}
	return strings.Join(lines, "\n")
}
