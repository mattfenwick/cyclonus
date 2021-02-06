package types

import "github.com/mattfenwick/cyclonus/pkg/utils"

type Table struct {
	Wrapped *utils.TruthTable
}

func NewTable(items []string) *Table {
	return &Table{Wrapped: utils.NewTruthTableFromItems(items, func() interface{} {
		return ConnectivityUnknown
	})}
}

func (r *Table) Set(from string, to string, value Connectivity) {
	r.Wrapped.Set(from, to, value)
}

func (r *Table) Get(from string, to string) Connectivity {
	return r.Wrapped.Get(from, to).(Connectivity)
}

func (r *Table) RenderTable() string {
	return r.Wrapped.Table(func(i interface{}) string {
		value := i.(Connectivity)
		return value.ShortString()
	})
}
