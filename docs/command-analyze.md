# cyclonus analyze

The `analyze` command provides a suite of tools for understanding and poking network policies,
without the need for a working kubernetes cluster or CNI.

It provides several modes which can be activated by a `--mode $MODE` flag.  Multiple `--mode` 
flags may be provided in a single invocation.

## Supported flags

```bash
cyclonus analyze -h
analyze network policies

Usage:
  cyclonus analyze [flags]

Flags:
  -A, --all-namespaces           reads kube resources from all namespaces; same as kubectl's '--all-namespaces'/'-A' flag
      --context string           selects kube context to read policies from; only reads from kube if one or more namespaces or all namespaces are specified
  -h, --help                     help for analyze
      --mode strings             analysis modes to run; allowed values are parse,explain,lint,query-traffic,query-target,probe (default [explain])
  -n, --namespace strings        namespaces to read kube resources from; similar to kubectl's '--namespace'/'-n' flag, except that multiple namespaces may be passed in and is empty if not set explicitly (instead of 'default' as in kubectl)
      --policy-path string       may be a file or a directory; if set, will attempt to read policies from the path
      --probe-path string        path to json model file for synthetic probe
      --simplify-policies        if true, reduce policies to simpler form while preserving semantics (default true)
      --target-pod-path string   path to json target pod file -- json array of dicts
      --traffic-path string      path to json traffic file, containing of a list of traffic objects
      --use-example-policies     if true, reads example policies

Global Flags:
  -v, --verbosity string   log level; one of [info, debug, trace, warn, error, fatal, panic] (default "info")
```

## Mode examples

### `--mode explain`: explains network policies

Groups policies by target, divides rules into egress and ingress, and gives a basic explanation of the combined
policies.  This clarifies the interactions between "denies" and "allows" from multiple policies.

```
cyclonus analyze \
  --mode explain \
  --policy-path ./networkpolicies/simple-example/

+---------+---------------+------------------------+---------------------+--------------------------+
|  TYPE   |    TARGET     |      SOURCE RULES      |        PEER         |      PORT/PROTOCOL       |
+---------+---------------+------------------------+---------------------+--------------------------+
| Ingress | namespace: y  | y/allow-label-to-label | no ips              | no ports, no protocols   |
|         | Match labels: | y/deny-all-for-label   |                     |                          |
|         |   pod: a      |                        |                     |                          |
+         +               +                        +---------------------+--------------------------+
|         |               |                        | namespace: y        | all ports, all protocols |
|         |               |                        | pods: Match labels: |                          |
|         |               |                        |   pod: c            |                          |
+         +---------------+------------------------+---------------------+                          +
|         | namespace: y  | y/allow-all-for-label  | all pods, all ips   |                          |
|         | Match labels: |                        |                     |                          |
|         |   pod: b      |                        |                     |                          |
+         +---------------+------------------------+---------------------+--------------------------+
|         | namespace: y  | y/allow-by-ip          | ports for all IPs   | no ports, no protocols   |
|         | Match labels: |                        |                     |                          |
|         |   pod: c      |                        |                     |                          |
+         +               +                        +---------------------+--------------------------+
|         |               |                        | 0.0.0.0/24          | all ports, all protocols |
|         |               |                        | except []           |                          |
|         |               |                        |                     |                          |
+         +               +                        +---------------------+--------------------------+
|         |               |                        | no pods             | no ports, no protocols   |
|         |               |                        |                     |                          |
|         |               |                        |                     |                          |
+         +---------------+------------------------+---------------------+                          +
|         | namespace: y  | y/deny-all             | no pods, no ips     |                          |
|         | all pods      |                        |                     |                          |
+---------+---------------+------------------------+---------------------+--------------------------+
```

### `--mode query-target`: which policy rules apply to a pod?

This takes the previous command a step further: it combines the rules from all the targets that apply
to a pod. 

```
cyclonus analyze \
  --mode query-target \
  --policy-path ./networkpolicies/simple-example/ \
  --target-pod-path ./examples/targets.json

pod in ns y with labels map[pod:a]:
+---------+---------------+-----------------------------+---------------------+--------------------------+
|  TYPE   |    TARGET     |        SOURCE RULES         |        PEER         |      PORT/PROTOCOL       |
+---------+---------------+-----------------------------+---------------------+--------------------------+
| Ingress | namespace: y  | y/allow-label-to-label      | no ips              | no ports, no protocols   |
|         | Match labels: | y/deny-all-for-label        |                     |                          |
|         |   pod: a      | y/deny-all                  |                     |                          |
+         +               +                             +---------------------+--------------------------+
|         |               |                             | namespace: y        | all ports, all protocols |
|         |               |                             | pods: Match labels: |                          |
|         |               |                             |   pod: c            |                          |
+---------+---------------+-----------------------------+---------------------+--------------------------+
|         |               |                             |                     |                          |
+---------+---------------+-----------------------------+---------------------+--------------------------+
| Egress  | namespace: y  | y/deny-all-egress           | all pods, all ips   | all ports, all protocols |
|         | Match labels: | y/allow-all-egress-by-label |                     |                          |
|         |   pod: a      |                             |                     |                          |
+---------+---------------+-----------------------------+---------------------+--------------------------+
```


### `--mode query-traffic`: will policies allow or block traffic?

Given arbitrary traffic examples (from a source to a destination, including labels, over a port and protocol),
this command parses network policies and determines if the traffic is allowed or not.

```
cyclonus analyze \
  --mode query-traffic \
  --policy-path ./networkpolicies/simple-example/ \
  --traffic-path ./examples/traffic.json

Traffic:
+--------------------------+-------------+---------------+-----------+-----------+------------+
|      PORT/PROTOCOL       | SOURCE/DEST |    POD IP     | NAMESPACE | NS LABELS | POD LABELS |
+--------------------------+-------------+---------------+-----------+-----------+------------+
| 80 (serve-80-tcp) on TCP | source      | 192.168.1.99  | y         | ns: y     | app: c     |
+                          +-------------+---------------+           +           +------------+
|                          | destination | 192.168.1.100 |           |           | pod: b     |
+--------------------------+-------------+---------------+-----------+-----------+------------+

Is traffic allowed?
+-------------+--------+---------------+
|    TYPE     | ACTION |    TARGET     |
+-------------+--------+---------------+
| Ingress     | Allow  | namespace: y  |
|             |        | Match labels: |
|             |        |   pod: b      |
+             +--------+---------------+
|             | Deny   | namespace: y  |
|             |        | all pods      |
+-------------+--------+---------------+
|             |        |               |
+-------------+--------+---------------+
| Egress      | Deny   | namespace: y  |
|             |        | all pods      |
+-------------+--------+---------------+
| IS ALLOWED? | FALSE  |                
+-------------+--------+---------------+
```

### `--mode probe`: simulates a connectivity probe

Runs a simulated connectivity probe against a set of network policies, without using a kubernetes cluster.

```
cyclonus analyze \
  --mode probe \
  --policy-path ./networkpolicies/simple-example/ \
  --probe-path ./examples/probe.json

Combined:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|     | X/A | X/B | X/C | Y/A | Y/B | Y/C | Z/A | Z/B | Z/C |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
| x/a | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| x/b | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| x/c | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| y/a | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| y/b | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| y/c | X   | X   | X   | X   | X   | X   | X   | X   | X   |
| z/a | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| z/b | .   | .   | .   | X   | .   | X   | .   | .   | .   |
| z/c | .   | .   | .   | X   | .   | X   | .   | .   | .   |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
```

### `--mode lint`: lints network policies

Checks network policies for common problems.

```
cyclonus analyze \
  --mode lint \
  --policy-path ./networkpolicies/simple-example

+-----------------+------------------------------+-------------------+-----------------------------+
| SOURCE/RESOLVED |             TYPE             |      TARGET       |       SOURCE POLICIES       |
+-----------------+------------------------------+-------------------+-----------------------------+
| Resolved        | CheckTargetAllEgressAllowed  | namespace: y      | y/allow-all-egress-by-label |
|                 |                              |                   |                             |
|                 |                              | pod selector:     |                             |
|                 |                              | matchExpressions: |                             |
|                 |                              | - key: pod        |                             |
|                 |                              |   operator: In    |                             |
|                 |                              |   values:         |                             |
|                 |                              |   - a             |                             |
|                 |                              |   - b             |                             |
|                 |                              |                   |                             |
+-----------------+------------------------------+-------------------+-----------------------------+
| Resolved        | CheckDNSBlockedOnTCP         | namespace: y      | y/deny-all-egress           |
|                 |                              |                   |                             |
|                 |                              | pod selector:     |                             |
|                 |                              | {}                |                             |
|                 |                              |                   |                             |
+-----------------+------------------------------+-------------------+-----------------------------+
| Resolved        | CheckDNSBlockedOnUDP         | namespace: y      | y/deny-all-egress           |
|                 |                              |                   |                             |
|                 |                              | pod selector:     |                             |
|                 |                              | {}                |                             |
|                 |                              |                   |                             |
+-----------------+------------------------------+-------------------+-----------------------------+
```