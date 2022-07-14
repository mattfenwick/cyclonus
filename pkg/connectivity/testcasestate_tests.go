package connectivity

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type buildLabelDiffCase struct {
	actual   map[string]string
	expected map[string]string
	diff     *LabelsDiff
}

type labelDiffMethodsCase struct {
	diff                        *LabelsDiff
	areLabelsEqual              bool
	areAllExpectedLabelsPresent bool
}

func RunTestCaseStateTests() {
	Describe("LabelDiff", func() {
		empty := map[string]string{}
		ab := map[string]string{"a": "b"}
		ac := map[string]string{"a": "c"}

		It("Build LabelDiff", func() {
			testCases := []*buildLabelDiffCase{
				{
					actual:   empty,
					expected: empty,
					diff:     &LabelsDiff{},
				},
				{
					actual:   empty,
					expected: ab,
					diff:     &LabelsDiff{Missing: []string{"a"}},
				},
				{
					actual:   ab,
					expected: empty,
					diff:     &LabelsDiff{Extra: []string{"a"}},
				},
				{
					actual:   ab,
					expected: ab,
					diff:     &LabelsDiff{Same: []string{"a"}},
				},
				{
					actual:   ab,
					expected: ac,
					diff:     &LabelsDiff{Different: []string{"a"}},
				},
				{
					actual:   ac,
					expected: ab,
					diff:     &LabelsDiff{Different: []string{"a"}},
				},
				{
					actual:   map[string]string{"a": "b", "c": "d"},
					expected: map[string]string{"c": "e", "a": "b"},
					diff:     &LabelsDiff{Same: []string{"a"}, Different: []string{"c"}},
				},
			}
			for _, tc := range testCases {
				diff := NewLabelsDiff(tc.actual, tc.expected)
				Expect(diff).To(Equal(tc.diff))
			}
		})

		It("Call LabelDiff methods", func() {
			testCases := []*labelDiffMethodsCase{
				{
					diff:                        &LabelsDiff{},
					areLabelsEqual:              true,
					areAllExpectedLabelsPresent: true,
				},
				{
					diff:                        &LabelsDiff{Same: []string{"c", "e"}},
					areLabelsEqual:              true,
					areAllExpectedLabelsPresent: true,
				},
				{
					diff:                        &LabelsDiff{Different: []string{"q"}},
					areLabelsEqual:              false,
					areAllExpectedLabelsPresent: false,
				},
				{
					diff:                        &LabelsDiff{Missing: []string{"z"}},
					areLabelsEqual:              false,
					areAllExpectedLabelsPresent: false,
				},
				{
					diff:                        &LabelsDiff{Extra: []string{"y"}},
					areLabelsEqual:              false,
					areAllExpectedLabelsPresent: true,
				},
			}
			for _, tc := range testCases {
				fmt.Printf("test case: %t, %t, %+v\n", tc.areLabelsEqual, tc.areAllExpectedLabelsPresent, tc.diff)
				Expect(tc.areLabelsEqual).To(Equal(tc.diff.AreLabelsEqual()))
				Expect(tc.areAllExpectedLabelsPresent).To(Equal(tc.diff.AreAllExpectedLabelsPresent()))
			}
		})
	})
}
