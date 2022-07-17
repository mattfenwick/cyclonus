package probe

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/builtins"
	"github.com/mattfenwick/collections/pkg/slices"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"sort"
	"strings"
)

func (r *Resources) RenderTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetHeader([]string{"Namespace", "NS Labels", "Pod", "Pod Labels", "IPs", "Containers/Ports"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	var nsSlice []string
	nsToPod := map[string][]*Pod{}
	for _, pod := range r.Pods {
		ns := pod.Namespace
		if _, ok := nsToPod[ns]; !ok {
			nsToPod[ns] = []*Pod{}
			nsSlice = append(nsSlice, ns)
		}
		if _, ok := r.Namespaces[ns]; !ok {
			panic(errors.Errorf("cannot handle pod %s/%s: namespace not found", ns, pod.Name))
		}
		nsToPod[ns] = append(nsToPod[ns], pod)
	}

	sort.Slice(nsSlice, func(i, j int) bool {
		return nsSlice[i] < nsSlice[j]
	})
	for _, ns := range nsSlice {
		labels := r.Namespaces[ns]
		nsLabelLines := labelsToLines(labels)
		for _, pod := range nsToPod[ns] {
			podLabelLines := labelsToLines(pod.Labels)
			for _, cont := range pod.Containers {
				table.Append([]string{
					ns,
					nsLabelLines,
					pod.Name,
					podLabelLines,
					fmt.Sprintf("pod: %s\nservice: %s", pod.IP, pod.ServiceIP),
					fmt.Sprintf("%s, port %s: %d on %s", cont.Name, cont.PortName, cont.Port, cont.Protocol),
				})
			}
		}
	}

	table.Render()
	return tableString.String()
}

func labelsToLines(labels map[string]string) string {
	keys := slices.SortBy(builtins.CompareOrdered[string], maps.Keys(labels))
	var lines []string
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, labels[key]))
	}
	return strings.Join(lines, "\n")
}
