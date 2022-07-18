package probe

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/slices"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"strings"
)

func (r *Resources) RenderTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetHeader([]string{"Namespace", "NS Labels", "Pod", "Pod Labels", "IPs", "Containers/Ports"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	nsToPod := map[string][]*Pod{}
	for _, pod := range r.Pods {
		ns := pod.Namespace
		if _, ok := nsToPod[ns]; !ok {
			nsToPod[ns] = []*Pod{}
		}
		if _, ok := r.Namespaces[ns]; !ok {
			panic(errors.Errorf("cannot handle pod %s/%s: namespace not found", ns, pod.Name))
		}
		nsToPod[ns] = append(nsToPod[ns], pod)
	}

	namespaces := slices.Sort(maps.Keys(nsToPod))
	for _, ns := range namespaces {
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
	keys := slices.Sort(maps.Keys(labels))
	var lines []string
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, labels[key]))
	}
	return strings.Join(lines, "\n")
}
