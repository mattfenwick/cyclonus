package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
)

type ComparisonTable struct {
	Wrapped *utils.TruthTable
}

func NewComparisonTable(items []string) *ComparisonTable {
	return &ComparisonTable{Wrapped: utils.NewTruthTableFromItems(items, nil)}
}

func NewComparisonTableFrom(kubeProbe *types.Table, simulatedProbe *types.Table) *ComparisonTable {
	if len(kubeProbe.Wrapped.Froms) != len(simulatedProbe.Wrapped.Froms) || len(kubeProbe.Wrapped.Tos) != len(simulatedProbe.Wrapped.Tos) {
		panic(errors.Errorf("cannot compare tables of different dimensions"))
	}
	for i, fr := range kubeProbe.Wrapped.Froms {
		if simulatedProbe.Wrapped.Froms[i] != fr {
			panic(errors.Errorf("cannot compare: from keys at index %d do not match (%s vs %s)", i, simulatedProbe.Wrapped.Froms[i], fr))
		}
	}
	for i, to := range kubeProbe.Wrapped.Tos {
		if simulatedProbe.Wrapped.Tos[i] != to {
			panic(errors.Errorf("cannot compare: to keys at index %d do not match (%s vs %s)", i, simulatedProbe.Wrapped.Tos[i], to))
		}
	}

	table := NewComparisonTable(kubeProbe.Wrapped.Froms)
	for _, key := range kubeProbe.Wrapped.Keys() {
		table.Set(key.From, key.To, equalsDict(kubeProbe.Get(key.From, key.To), simulatedProbe.Get(key.From, key.To)))
	}

	return table
}

func equalsDict(l map[string]types.Connectivity, r map[string]types.Connectivity) bool {
	if len(l) != len(r) {
		return false
	}
	for k, lv := range l {
		if rv, ok := r[k]; !ok || rv != lv {
			return false
		}
	}
	return true
}

func (c *ComparisonTable) Set(from string, to string, value bool) {
	c.Wrapped.Set(from, to, value)
}

func (c *ComparisonTable) Get(from string, to string) bool {
	return c.Wrapped.Get(from, to).(bool)
}

func (c *ComparisonTable) ValueCounts(ignoreLoopback bool) map[Comparison]int {
	counts := map[Comparison]int{}
	for _, key := range c.Wrapped.Keys() {
		if ignoreLoopback && key.From == key.To {
			counts[IgnoredComparison] += 1
		} else {
			if c.Get(key.From, key.To) {
				counts[SameComparison] += 1
			} else {
				counts[DifferentComparison] += 1
			}
		}
	}
	return counts
}

func (c *ComparisonTable) RenderTable() string {
	return c.Wrapped.Table("", false, func(i interface{}) string {
		if i.(bool) {
			return "."
		} else {
			return "X"
		}
	})
}

type Comparison string

const (
	SameComparison      Comparison = "same"
	DifferentComparison Comparison = "different"
	IgnoredComparison   Comparison = "ignored"
)

func (c Comparison) ShortString() string {
	switch c {
	case SameComparison:
		return "."
	case DifferentComparison:
		return "X"
	case IgnoredComparison:
		return "?"
	default:
		panic(errors.Errorf("invalid Comparison value %+v", c))
	}
}
