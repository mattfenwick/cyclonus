package synthetic

import (
	"github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	"github.com/mattfenwick/cyclonus/pkg/utils"
)

type ResultTable struct {
	Wrapped *utils.TruthTable
}

func NewResultTable(items []string) *ResultTable {
	return &ResultTable{Wrapped: utils.NewTruthTableFromItems(items, func() interface{} {
		return &types.Answer{Ingress: types.ConnectivityUnknown, Egress: types.ConnectivityUnknown}
	})}
}

func (r *ResultTable) SetIngress(from string, to string, value types.Connectivity) {
	answer := r.Get(from, to)
	answer.Ingress = value
}

func (r *ResultTable) SetEgress(from string, to string, value types.Connectivity) {
	answer := r.Get(from, to)
	answer.Egress = value
}

func (r *ResultTable) Get(from string, to string) *types.Answer {
	value := r.Wrapped.Get(from, to)
	return value.(*types.Answer)
}

func (r *ResultTable) Table() string {
	return r.Wrapped.Table(func(i interface{}) string {
		value := i.(*types.Answer)
		return value.ShortString()
	})
}
