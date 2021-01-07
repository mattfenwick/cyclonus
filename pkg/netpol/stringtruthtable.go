package netpol

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"os"
)

type StringTruthTable struct {
	Froms  []string
	Tos    []string
	toSet  map[string]bool
	Values map[string]map[string]string
}

// constructors

func NewStringTruthTable(items []string) *StringTruthTable {
	return newStringTruthTable(items, items, nil)
}

func NewStringTruthTableWithFromsTo(froms []string, tos []string) *StringTruthTable {
	return newStringTruthTable(froms, tos, nil)
}

func NewStringTruthTableWithDefaultValue(items []string, defaultValue string) *StringTruthTable {
	return newStringTruthTable(items, items, &defaultValue)
}

func newStringTruthTable(froms []string, tos []string, defaultValue *string) *StringTruthTable {
	values := map[string]map[string]string{}
	for _, from := range froms {
		values[from] = map[string]string{}
		if defaultValue != nil {
			for _, to := range tos {
				values[from][to] = *defaultValue
			}
		}
	}
	toSet := map[string]bool{}
	for _, to := range tos {
		toSet[to] = true
	}
	return &StringTruthTable{
		Froms:  froms,
		Tos:    tos,
		toSet:  toSet,
		Values: values,
	}
}

// methods

// IsComplete returns true if there's a value set for every single pair of items, otherwise it returns false.
func (tt *StringTruthTable) IsComplete() bool {
	for _, from := range tt.Froms {
		for _, to := range tt.Tos {
			if _, ok := tt.Values[from][to]; !ok {
				return false
			}
		}
	}
	return true
}

func (tt *StringTruthTable) Set(from string, to string, value string) {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.Errorf("from-key %s not found", from))
	}
	if _, ok := tt.toSet[to]; !ok {
		panic(errors.Errorf("to-key %s not allowed", to))
	}
	dict[to] = value
}

func (tt *StringTruthTable) SetAllFrom(from string, value string) {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.Errorf("from-key %s not found", from))
	}
	for _, to := range tt.Tos {
		dict[to] = value
	}
}

func (tt *StringTruthTable) SetAllTo(to string, value string) {
	if _, ok := tt.toSet[to]; !ok {
		panic(errors.Errorf("to-key %s not found", to))
	}
	for _, from := range tt.Froms {
		tt.Values[from][to] = value
	}
}

func (tt *StringTruthTable) Get(from string, to string) string {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.Errorf("from-key %s not found", from))
	}
	val, ok := dict[to]
	if !ok {
		panic(errors.Errorf("to-key %s not found in map (%+v)", to, dict))
	}
	return val
}

func (tt *StringTruthTable) Compare(other *StringTruthTable) *StringTruthTable {
	if len(tt.Froms) != len(other.Froms) || len(tt.Tos) != len(other.Tos) {
		panic("cannot compare tables of different dimensions")
	}
	for i, fr := range tt.Froms {
		if other.Froms[i] != fr {
			panic(errors.Errorf("cannot compare: from keys at index %d do not match (%s vs %s)", i, other.Froms[i], fr))
		}
	}
	for i, to := range tt.Tos {
		if other.Tos[i] != to {
			panic(errors.Errorf("cannot compare: to keys at index %d do not match (%s vs %s)", i, other.Tos[i], to))
		}
	}

	values := map[string]map[string]string{}
	for from, dict := range tt.Values {
		values[from] = map[string]string{}
		for to, val := range dict {
			values[from][to] = fmt.Sprintf("%t", val == other.Values[from][to]) // TODO other.Get(from, to) ?
		}
	}
	// TODO check for equality from both sides
	return &StringTruthTable{
		Froms:  tt.Froms,
		Tos:    tt.Tos,
		toSet:  tt.toSet,
		Values: values,
	}
}

func (tt *StringTruthTable) Table() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(append([]string{"-"}, tt.Tos...))

	for _, from := range tt.Froms {
		line := []string{from}
		for _, to := range tt.Tos {
			val, ok := tt.Values[from][to]
			if ok {
				line = append(line, val)
			} else {
				line = append(line, "?")
			}
		}
		table.Append(line)
	}

	return table
}
