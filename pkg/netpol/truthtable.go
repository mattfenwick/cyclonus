package netpol

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type TruthTable struct {
	Items   []string
	itemSet map[string]bool
	Values  map[string]map[string]bool
}

func NewTruthTable(items []string, defaultValue *bool) *TruthTable {
	itemSet := map[string]bool{}
	values := map[string]map[string]bool{}
	for _, from := range items {
		itemSet[from] = true
		values[from] = map[string]bool{}
		if defaultValue != nil {
			for _, to := range items {
				values[from][to] = *defaultValue
			}
		}
	}
	return &TruthTable{
		Items:   items,
		itemSet: itemSet,
		Values:  values,
	}
}

// IsComplete returns true if there's a value set for every single pair of items, otherwise it returns false.
func (tt *TruthTable) IsComplete() bool {
	for _, from := range tt.Items {
		for _, to := range tt.Items {
			if _, ok := tt.Values[from][to]; !ok {
				return false
			}
		}
	}
	return true
}

func (tt *TruthTable) Set(from string, to string, value bool) {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.New(fmt.Sprintf("key %s not found in map", from)))
	}
	if _, ok := tt.itemSet[to]; !ok {
		panic(errors.New(fmt.Sprintf("key %s not allowed", to)))
	}
	dict[to] = value
}

func (tt *TruthTable) SetAllFrom(from string, value bool) {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.New(fmt.Sprintf("key %s not found in map", from)))
	}
	for _, to := range tt.Items {
		dict[to] = value
	}
}

func (tt *TruthTable) SetAllTo(to string, value bool) {
	if _, ok := tt.itemSet[to]; !ok {
		panic(errors.New(fmt.Sprintf("key %s not found", to)))
	}
	for _, from := range tt.Items {
		tt.Values[from][to] = value
	}
}

func (tt *TruthTable) Get(from string, to string) bool {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.New(fmt.Sprintf("key %s not found in map", from)))
	}
	val, ok := dict[to]
	if !ok {
		panic(errors.New(fmt.Sprintf("key %s not found in map (%+v)", to, dict)))
	}
	return val
}

func (tt *TruthTable) Compare(other *TruthTable) *TruthTable {
	// TODO set equality
	//if tt.itemSet != other.itemSet {
	//	panic()
	//}
	values := map[string]map[string]bool{}
	for from, dict := range tt.Values {
		values[from] = map[string]bool{}
		for to, val := range dict {
			values[from][to] = val == other.Values[from][to] // TODO other.Get(from, to) ?
		}
	}
	// TODO check for equality from both sides
	return &TruthTable{
		Items:   tt.Items,
		itemSet: tt.itemSet,
		Values:  values,
	}
}

func (tt *TruthTable) PrettyPrint() string {
	header := strings.Join(append([]string{"-"}, tt.Items...), "\t")
	lines := []string{header}
	for _, from := range tt.Items {
		line := []string{from}
		for _, to := range tt.Items {
			val := "X"
			if tt.Values[from][to] {
				val = "."
			}
			line = append(line, val)
		}
		lines = append(lines, strings.Join(line, "\t"))
	}
	return strings.Join(lines, "\n")
}

func (tt *TruthTable) Table() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(append([]string{"-"}, tt.Items...))

	for _, from := range tt.Items {
		line := []string{from}
		for _, to := range tt.Items {
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
