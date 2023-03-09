# cyclonus generate

For CNI conformance testing.

Generate network policy test scenarios, install the scenarios one at a time in kubernetes,
and compare actual measured connectivity to expected connectivity using a truth table.

## Supported flags

```bash
generate network policies, create and probe against kubernetes, and compare to expected results

Usage:
  cyclonus generate [flags]

Flags:
      --allow-dns                          if using egress, allow tcp and udp over port 53 for DNS resolution (default true)
      --cleanup-namespaces                 if true, clean up namespaces after completion
      --context string                     kubernetes context to use; if empty, uses default context
      --destination-type string            override to set what to direct requests at; if not specified, the tests will be left as-is; one of service-name, service-ip, pod-ip
      --dry-run                            if true, don't actually do anything: just print out what would be done
      --exclude strings                    exclude tests with any of these tags.  See 'include' field for valid tags (default [multi-peer,upstream-e2e,example,end-port])
  -h, --help                               help for generate
      --ignore-loopback                    if true, ignore loopback for truthtable correctness verification
      --include strings                    include tests with any of these tags; if empty, all tests will be included.
      --job-timeout-seconds int            number of seconds to pass on to 'agnhost connect --timeout=%ds' flag (default 10)
      --junit-results-file string          output junit results to the specified file
      --mock                               if true, use a mock kube runner (i.e. don't actually run tests against kubernetes; instead, product fake results
      --namespace strings                  namespaces to create/use pods in (default [x,y,z])
      --noisy                              if true, print all results
      --perturbation-wait-seconds int      number of seconds to wait after perturbing the cluster (i.e. create a network policy, modify a ns/pod label) before running probes, to give the CNI time to update the cluster state (default 5)
      --pod strings                        pods to create in namespaces (default [a,b,c])
      --pod-creation-timeout-seconds int   number of seconds to wait for pods to create, be running and have IP addresses (default 60)
      --retries int                        number of kube probe retries to allow, if probe fails (default 1)
      --server-port ints                   ports to run server on (default [80,81])
      --server-protocol strings            protocols to run server on (default [TCP,UDP,SCTP])

Global Flags:
  -v, --verbosity string   log level; one of [info, debug, trace, warn, error, fatal, panic] (default "info")
```

## Example

```
cyclonus generate \
  --mode simple-fragments \
  --include conflict,peer-ipblock \
  --ignore-loopback \
  --perturbation-wait-seconds 15

...
Tag results:
```
| Tag | Result |
| --- | --- |
| direction | 10 / 20 = 50% ❌ |
|  - egress | 5 / 11 = 45% ❌ |
|  - ingress | 5 / 11 = 45% ❌ |
| miscellaneous | 10 / 16 = 62% ❌ |
|  - conflict | 10 / 16 = 62% ❌ |
| peer-ipblock | 0 / 4 = 0% ❌ |
|  - IP-block-no-except | 0 / 2 = 0% ❌ |
|  - IP-block-with-except | 0 / 2 = 0% ❌ |
| peer-pods | 4 / 4 = 100% ✅ |
|  - all-namespaces | 4 / 4 = 100% ✅ |
|  - all-pods | 4 / 4 = 100% ✅ |
| rule | 6 / 8 = 75% ❌ |
|  - allow-all | 2 / 4 = 50% ❌ |
|  - deny-all | 6 / 8 = 75% ❌ |