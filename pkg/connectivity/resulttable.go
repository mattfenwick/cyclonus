package connectivity

import (
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
)

type Combined struct {
	Kube      types.Connectivity
	Synthetic *types.Answer
}

type ResultTable struct {
	Wrapped *utils.TruthTable
}

func NewResultTable(items []string) *ResultTable {
	return &ResultTable{Wrapped: utils.NewTruthTableFromItems(items, nil)}
}

func NewResultTableFrom(kubeResults *connectivitykube.ResultTable, syntheticResults *synthetic.ResultTable) *ResultTable {
	if len(kubeResults.Wrapped.Froms) != len(syntheticResults.Wrapped.Froms) || len(kubeResults.Wrapped.Tos) != len(syntheticResults.Wrapped.Tos) {
		panic(errors.Errorf("cannot compare tables of different dimensions"))
	}
	for i, fr := range kubeResults.Wrapped.Froms {
		if syntheticResults.Wrapped.Froms[i] != fr {
			panic(errors.Errorf("cannot compare: from keys at index %d do not match (%s vs %s)", i, syntheticResults.Wrapped.Froms[i], fr))
		}
	}
	for i, to := range kubeResults.Wrapped.Tos {
		if syntheticResults.Wrapped.Tos[i] != to {
			panic(errors.Errorf("cannot compare: to keys at index %d do not match (%s vs %s)", i, syntheticResults.Wrapped.Tos[i], to))
		}
	}

	table := NewResultTable(kubeResults.Wrapped.Froms)

	for _, key := range kubeResults.Wrapped.Keys() {
		table.Set(key.From, key.To, &Combined{
			Kube:      kubeResults.Get(key.From, key.To),
			Synthetic: syntheticResults.Get(key.From, key.To),
		})
	}

	return table
}

func (r *ResultTable) Set(from string, to string, value *Combined) {
	r.Wrapped.Set(from, to, value)
}

func (r *ResultTable) Get(from string, to string) *Combined {
	return r.Wrapped.Get(from, to).(*Combined)
}

type Summary struct {
	Counts map[Comparison]int
}

func (r *ResultTable) ValueCounts(ignoreLoopback bool) *Summary {
	s := &Summary{Counts: map[Comparison]int{}}
	for _, key := range r.Wrapped.Keys() {
		if ignoreLoopback && key.From == key.To {
			s.Counts[IgnoredComparison] += 1
		} else {
			elem := r.Get(key.From, key.To)
			if elem.Kube == elem.Synthetic.Combined() {
				s.Counts[SameComparison] += 1
			} else {
				s.Counts[DifferentComparison] += 1
			}
		}
	}
	return s
}

type Comparison string

const (
	SameComparison      Comparison = "same"
	DifferentComparison Comparison = "different"
	IgnoredComparison   Comparison = "ignored"
)
