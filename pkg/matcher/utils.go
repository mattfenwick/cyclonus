package matcher

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

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
