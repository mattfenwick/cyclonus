package kube

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"sort"
	"strings"
)

func (r *Resources) Table() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetHeader([]string{"Namespace", "NS Labels", "Pod", "Pod Labels", "Container", "Ports"})
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
			for _, cont := range pod.KubeContainers() {
				var ports []string
				for _, p := range cont.Ports {
					ports = append(ports, fmt.Sprintf("%s: %d on %s", p.Name, p.ContainerPort, p.Protocol))
				}
				table.Append([]string{
					ns,
					labelsToLines(labels),
					pod.Name,
					labelsToLines(pod.Labels),
					cont.Name,
					//strings.Join(cont.Command, " "),
					strings.Join(ports, "\n"),
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
