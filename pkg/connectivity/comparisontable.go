package connectivity

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type Item struct {
	Kube      *probe.Item
	Simulated *probe.Item
}

func (i *Item) ResultsByProtocol() map[bool]map[v1.Protocol]int {
	counts := map[bool]map[v1.Protocol]int{true: {}, false: {}}
	for key, kr := range i.Kube.JobResults {
		counts[kr.Combined == i.Simulated.JobResults[key].Combined][kr.Job.Protocol]++
	}
	return counts
}

func (i *Item) IsSuccess() bool {
	return equalsDict(i.Kube.JobResults, i.Simulated.JobResults)
}

func equalsDict(l map[string]*probe.JobResult, r map[string]*probe.JobResult) bool {
	if len(l) != len(r) {
		return false
	}
	for k, lv := range l {
		if rv, ok := r[k]; !ok || rv.Combined != lv.Combined {
			return false
		}
	}
	return true
}

type ComparisonTable struct {
	Wrapped *probe.TruthTable
}

func NewComparisonTable(items []string) *ComparisonTable {
	return &ComparisonTable{Wrapped: probe.NewTruthTableFromItems(items, nil)}
}

func NewComparisonTableFrom(kubeProbe *probe.Table, simulatedProbe *probe.Table) *ComparisonTable {
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
		table.Set(key.From, key.To, &Item{Kube: kubeProbe.Get(key.From, key.To), Simulated: simulatedProbe.Get(key.From, key.To)})
	}

	return table
}

func (c *ComparisonTable) ResultsByProtocol() map[bool]map[v1.Protocol]int {
	counts := map[bool]map[v1.Protocol]int{true: {}, false: {}}
	for _, key := range c.Wrapped.Keys() {
		for isSuccess, protocolCounts := range c.Get(key.From, key.To).ResultsByProtocol() {
			for protocol, count := range protocolCounts {
				counts[isSuccess][protocol] += count
			}
		}
	}
	return counts
}

func (c *ComparisonTable) Set(from string, to string, value *Item) {
	c.Wrapped.Set(from, to, value)
}

func (c *ComparisonTable) Get(from string, to string) *Item {
	return c.Wrapped.Get(from, to).(*Item)
}

func (c *ComparisonTable) ValueCounts(ignoreLoopback bool) map[Comparison]int {
	counts := map[Comparison]int{}
	for _, key := range c.Wrapped.Keys() {
		if ignoreLoopback && key.From == key.To {
			counts[IgnoredComparison] += 1
		} else {
			if c.Get(key.From, key.To).IsSuccess() {
				counts[SameComparison] += 1
			} else {
				counts[DifferentComparison] += 1
			}
		}
	}
	return counts
}

func (c *ComparisonTable) RenderSuccessTable() string {
	return c.Wrapped.Table("", false, func(fr, to string, i interface{}) string {
		item := c.Get(fr, to)
		if item.IsSuccess() {
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
