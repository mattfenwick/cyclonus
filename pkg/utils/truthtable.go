package utils

import (
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"strings"
)

type TableKey struct {
	From string
	To   string
}

// TruthTable takes in n items and maintains an n x n table of booleans for each ordered pair
type TruthTable struct {
	Froms  []string
	Tos    []string
	toSet  map[string]bool
	Values map[string]map[string]interface{}
}

// NewTruthTableFromItems creates a new truth table with items
func NewTruthTableFromItems(items []string, defaultValue func() interface{}) *TruthTable {
	return NewTruthTable(items, items, defaultValue)
}

// NewTruthTable creates a new truth table with froms and tos
func NewTruthTable(froms []string, tos []string, defaultValue func() interface{}) *TruthTable {
	values := map[string]map[string]interface{}{}
	for _, from := range froms {
		values[from] = map[string]interface{}{}
		for _, to := range tos {
			if defaultValue != nil {
				values[from][to] = defaultValue()
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
func (tt *TruthTable) Set(from string, to string, value interface{}) {
	dict, ok := tt.Values[from]
	if !ok {
		panic(errors.Errorf("from-key %s not found", from))
	}
	if _, ok := tt.toSet[to]; !ok {
		panic(errors.Errorf("to-key %s not allowed", to))
	}
	dict[to] = value
}

// Get gets the specified value
func (tt *TruthTable) Get(from string, to string) interface{} {
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

func (tt *TruthTable) GetKey(key *TableKey) interface{} {
	return tt.Get(key.From, key.To)
}

func (tt *TruthTable) Keys() []*TableKey {
	var keys []*TableKey
	for _, from := range tt.Froms {
		for _, to := range tt.Tos {
			keys = append(keys, &TableKey{From: from, To: to})
		}
	}
	return keys
}

func (tt *TruthTable) Table(printElement func(interface{}) string) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader(append([]string{"-"}, tt.Tos...))

	for _, from := range tt.Froms {
		line := []string{from}
		for _, to := range tt.Tos {
			line = append(line, printElement(tt.Values[from][to]))
		}
		table.Append(line)
	}

	table.Render()
	return tableString.String()
}
