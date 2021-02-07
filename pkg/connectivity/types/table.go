package types

import (
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"sort"
	"strings"
)

type Table struct {
	Wrapped *utils.TruthTable
}

func NewTable(items []string) *Table {
	return &Table{Wrapped: utils.NewTruthTableFromItems(items, func() interface{} {
		return map[string]Connectivity{}
	})}
}

func (r *Table) Set(from string, to string, key string, value Connectivity) {
	dict := r.Get(from, to)
	dict[key] = value
}

func (r *Table) Get(from string, to string) map[string]Connectivity {
	return r.Wrapped.Get(from, to).(map[string]Connectivity)
}

func (r *Table) RenderTable() string {
	isSchemaUniform, isSingleElement := true, true
	schema := map[string]bool{}
Outer:
	for _, fr := range r.Wrapped.Froms {
		for _, to := range r.Wrapped.Tos {
			dict := r.Get(fr, to)
			if len(dict) != 1 {
				isSingleElement = false
				break Outer
			}
			var keys []string
			for k := range dict {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			schema[strings.Join(keys, "_")] = true
			if len(schema) > 1 {
				isSchemaUniform = false
				break Outer
			}
		}
	}
	if isSchemaUniform && isSingleElement {
		return r.singleElementRenderTable()
	} else {
		return r.multiElementRenderTable()
	}
}

func (r *Table) singleElementRenderTable() string {
	return r.Wrapped.Table(false, func(i interface{}) string {
		dict := i.(map[string]Connectivity)
		var v Connectivity
		for _, value := range dict {
			v = value
			break
		}
		return v.ShortString()
	})
}

func (r *Table) multiElementRenderTable() string {
	return r.Wrapped.Table(true, func(i interface{}) string {
		dict := i.(map[string]Connectivity)
		var keys []string
		for k := range dict {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var lines []string
		for _, k := range keys {
			lines = append(lines, k+": "+dict[k].ShortString())
		}
		return strings.Join(lines, "\n")
	})
}
