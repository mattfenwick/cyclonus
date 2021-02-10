package probe

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"sort"
	"strings"
)

func (r *Resources) RenderTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetHeader([]string{"Namespace", "NS Labels", "Pod", "Pod Labels", "Pod IP", "Containers/Ports"})
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
		for _, pod := range nsToPod[ns] {
			for _, cont := range pod.Containers {
				table.Append([]string{
					ns,
					labelsToLines(labels),
					pod.Name,
					labelsToLines(pod.Labels),
					pod.IP,
					fmt.Sprintf("%s, port %s: %d on %s", cont.Name, cont.PortName, cont.Port, cont.Protocol),
				})
			}
		}
	}

	table.Render()
	return tableString.String()
}

func labelsToLines(labels map[string]string) string {
	var lines []string
	for k, v := range labels {
		lines = append(lines, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(lines, "\n")
}
