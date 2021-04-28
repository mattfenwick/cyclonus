# cyclonus probe

Run a connectivity probe against a Kubernetes cluster.

## Supported flags

```bash
cyclonus probe -h
run a connectivity probe against kubernetes pods

Usage:
  cyclonus probe [flags]

Flags:
      --all-available                      if true, probe all available ports and protocols on each pod (default true)
      --context string                     kubernetes context to use; if empty, uses default context
  -h, --help                               help for probe
      --ignore-loopback                    if true, ignore loopback for truthtable correctness verification
      --job-timeout-seconds int            number of seconds to pass on to 'agnhost connect --timeout=%ds' flag (default 10)
      --noisy                              if true, print all results
      --perturbation-wait-seconds int      number of seconds to wait after perturbing the cluster (i.e. create a network policy, modify a ns/pod label) before running probes, to give the CNI time to update the cluster state (default 5)
      --pod-creation-timeout-seconds int   number of seconds to wait for pods to create, be running and have IP addresses (default 60)
      --policy-path string                 path to yaml network policy to create in kube; if empty, will not create any policies
      --port strings                       ports to run probes on; may be named port or numbered port (default [80])
      --probe-mode string                  probe mode to use, must be one of service-name, service-ip, pod-ip (default "service-name")
      --protocol strings                   protocols to run probes on (default [tcp])
  -n, --server-namespace strings           namespaces to create/use pods in (default [x,y,z])
      --server-pod strings                 pods to create in namespaces (default [a,b,c])
      --server-port ints                   ports to run server on (default [80,81])
      --server-protocol strings            protocols to run server on (default [TCP,UDP,SCTP])

Global Flags:
  -v, --verbosity string   log level; one of [info, debug, trace, warn, error, fatal, panic] (default "info")
```

## Example

```
cyclonus probe

Kube results for:
  policy y/allow-all-for-label:
  policy y/allow-by-ip:
  policy y/allow-label-to-label:
  policy y/deny-all:
  policy y/deny-all-for-label:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|  -  | X/A | X/B | X/C | Y/A | Y/B | Y/C | Z/A | Z/B | Z/C |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
| x/a | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| x/b | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| x/c | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| y/a | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| y/b | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| y/c | .   | .   | .   | .   | .   | X   | .   | .   | .   |
| z/a | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| z/b | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| z/c | .   | .   | .   | X   | .   | X   | .   | .   | .   |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+

0 wrong, 0 no value, 81 correct, 0 ignored out of 81 total
```