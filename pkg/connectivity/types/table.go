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

	for _, key := range r.Wrapped.Keys() {
		dict := r.Get(key.From, key.To)
		if len(dict) != 1 {
			isSingleElement = false
			break
		}
		var keys []string
		for k := range dict {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		schema[strings.Join(keys, "_")] = true
		if len(schema) > 1 {
			isSchemaUniform = false
			break
		}
	}
	if isSchemaUniform && isSingleElement {
		return r.renderSimpleTable()
	} else if isSchemaUniform {
		return r.renderUniformMultiTable()
	} else {
		return r.renderNonuniformTable()
	}
}

func (r *Table) renderSimpleTable() string {
	return r.Wrapped.Table("", false, func(i interface{}) string {
		dict := i.(map[string]Connectivity)
		var v Connectivity
		for _, value := range dict {
			v = value
			break
		}
		return v.ShortString()
	})
}

func (r *Table) renderUniformMultiTable() string {
	key := r.Wrapped.Keys()[0]
	first := r.Get(key.From, key.To)
	var keys []string
	for k := range first {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	schema := strings.Join(keys, "\n")
	return r.Wrapped.Table(schema, true, func(i interface{}) string {
		dict := i.(map[string]Connectivity)
		var lines []string
		for _, k := range keys {
			lines = append(lines, dict[k].ShortString())
		}
		return strings.Join(lines, "\n")
	})
}

func (r *Table) renderNonuniformTable() string {
	return r.Wrapped.Table("", true, func(i interface{}) string {
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
