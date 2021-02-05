package kube

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/utils"
)

type ResultTable struct {
	Wrapped *utils.TruthTable
}

func NewResultTable(items []string) *ResultTable {
	return &ResultTable{Wrapped: utils.NewTruthTableFromItems(items, func() interface{} {
		return types.ConnectivityUnknown
	})}
}

func (r *ResultTable) Set(from string, to string, value types.Connectivity) {
	r.Wrapped.Set(from, to, value)
}

func (r *ResultTable) Get(from string, to string) types.Connectivity {
	value := r.Wrapped.Get(from, to)
	return value.(types.Connectivity)
}

func (r *ResultTable) Table() string {
	return r.Wrapped.Table(func(i interface{}) string {
		value := i.(types.Connectivity)
		return value.ShortString()
	})
}
