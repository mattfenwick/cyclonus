package utils

import (
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"os"
)

// TruthTable takes in n items and maintains an n x n table of booleans for each ordered pair
type TruthTable struct {
	Froms  []string
	Tos    []string
	toSet  map[string]bool
	Values map[string]map[string]bool
}

// NewTruthTableFromItems creates a new truth table with items
func NewTruthTableFromItems(items []string, defaultValue *bool) *TruthTable {
	return NewTruthTable(items, items, defaultValue)
}

// NewTruthTable creates a new truth table with froms and tos
func NewTruthTable(froms []string, tos []string, defaultValue *bool) *TruthTable {
	values := map[string]map[string]bool{}
	for _, from := range froms {
		values[from] = map[string]bool{}
		for _, to := range tos {
			if defaultValue != nil {
				values[from][to] = *defaultValue
			}
		}
	}
	toSet := map[string]bool{}
	for _, to := range tos {
		toSet[to] = true
	}
	return &TruthTable{
		Froms:  froms,
		Tos:    tos,
		toSet:  toSet,
		Values: values,
	}
}

// IsComplete returns true if there's a value set for every single pair of items, otherwise it returns false.
func (tt *TruthTable) IsComplete() bool {
	for _, from := range tt.Froms {
		for _, to := range tt.Tos {
			if _, ok := tt.Values[from][to]; !ok {
				return false
			}
		}
	}
	return true
}

// Set sets the value for from->to
func (tt *TruthTable) Set(from string, to string, value bool) {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.Errorf("from-key %s not found", from))
	}
	if _, ok := tt.toSet[to]; !ok {
		panic(errors.Errorf("to-key %s not allowed", to))
	}
	dict[to] = value
}

// SetAllFrom sets all values where from = 'from'
func (tt *TruthTable) SetAllFrom(from string, value bool) {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.Errorf("from-key %s not found", from))
	}
	for _, to := range tt.Tos {
		dict[to] = value
	}
}

// SetAllTo sets all values where to = 'to'
func (tt *TruthTable) SetAllTo(to string, value bool) {
	if _, ok := tt.toSet[to]; !ok {
		panic(errors.Errorf("to-key %s not found", to))
	}
	for _, from := range tt.Froms {
		tt.Values[from][to] = value
	}
}

// Get gets the specified value
func (tt *TruthTable) Get(from string, to string) bool {
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

func (tt *TruthTable) ValueCounts(ignoreLoopback bool) (int, int, int, int) {
	trueCount, falseCount, noValueCount, totalChecked := 0, 0, 0, 0
	for _, from := range tt.Froms {
		for _, to := range tt.Tos {
			if ignoreLoopback && from == to {
				continue
			}
			if _, ok := tt.Values[from][to]; !ok {
				noValueCount++
			} else if tt.Values[from][to] {
				trueCount++
			} else {
				falseCount++
			}
			totalChecked++
		}
	}
	return trueCount, falseCount, noValueCount, totalChecked
}

// Compare is used to check two truth tables for equality, returning its
// result in the form of a third truth table.  Both tables are expected to
// have identical items.
func (tt *TruthTable) Compare(other *TruthTable) *TruthTable {
	if len(tt.Froms) != len(other.Froms) || len(tt.Tos) != len(other.Tos) {
		panic(errors.Errorf("cannot compare tables of different dimensions"))
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

	values := map[string]map[string]bool{}
	for from, dict := range tt.Values {
		values[from] = map[string]bool{}
		for to, val := range dict {
			values[from][to] = val == other.Values[from][to]
		}
	}
	return &TruthTable{
		Froms:  tt.Froms,
		Tos:    tt.Tos,
		toSet:  tt.toSet,
		Values: values,
	}
}

// PrettyPrint produces a nice visual representation.
//func (tt *TruthTable) PrettyPrint(indent string) string {
//	header := indent + strings.Join(append([]string{"-\t"}, tt.Tos...), "\t")
//	lines := []string{header}
//	for _, from := range tt.Froms {
//		line := []string{from}
//		for _, to := range tt.Tos {
//			mark := "X"
//			val, ok := tt.Values[from][to]
//			if !ok {
//				mark = "?"
//			} else if val {
//				mark = "."
//			}
//			line = append(line, mark+"\t")
//		}
//		lines = append(lines, indent+strings.Join(line, "\t"))
//	}
//	return strings.Join(lines, "\n")
//}

func (tt *TruthTable) Table() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(append([]string{"-"}, tt.Tos...))

	for _, from := range tt.Froms {
		line := []string{from}
		for _, to := range tt.Tos {
			val := "?"
			isTrue, ok := tt.Values[from][to]
			if isTrue {
				val = "."
			} else if ok {
				val = "X"
			}
			line = append(line, val)
		}
		table.Append(line)
	}

	return table
}
