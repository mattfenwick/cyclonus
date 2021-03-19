# Cyclonus

## Network policy explainer, prober, and test case generator

Parse, explain, and probe network policies to understand their implications and help design
policies that suit your needs!

### Quickstart

Grab the [latest release](https://github.com/mattfenwick/cyclonus/releases) to get started using Cyclonus.

### Run as a kubernetes job

Take a look at the [hack directory](./hack):

```
ls hack
antrea              calico              cilium              flannel             run-cyclonus-job.sh
```

Choose the right job for your CNI, then run:

```
cd hack/calico
./github-action.sh
```

This will:

 - create a namespace, service account, and cluster role binding for cyclonus
 - create a cyclonus job

Pull the logs from the job:

```
kubectl logs -f -n netpol cyclonus-abcde
```

## Docker images

Images are available at [mfenwick100/cyclonus](https://hub.docker.com/r/mfenwick100/cyclonus/tags?page=1&ordering=last_updated):

```
docker pull docker.io/mfenwick100/cyclonus:latest
```

## Integrations

### krew plugin

Cyclonus is available as a [krew/kubectl plugin](https://github.com/mattfenwick/kubectl-cyclonus):

 - [Set up krew](https://krew.sigs.k8s.io/docs/user-guide/quickstart/)
 - install cyclonus through krew: `kubectl krew install cyclonus`
 - use cyclonus as a kubectl plugin: `kubectl cyclonus -h`.

### Antrea testing

Cyclonus runs network policy tests for Antrea on a daily basis;
[check it out on github](https://github.com/vmware-tanzu/antrea/actions/workflows/netpol_cyclonus.yml).

### Cilium testing (in progress)

[Cyclonus will run network policy tests for Cilium on a daily basis](https://github.com/cilium/cilium/pull/14889).

## Probe

Run a connectivity probe against a Kubernetes cluster.

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

## Policy generator

For CNI conformance testing.

Generate network policy test scenarios, install the scenarios one at a time in kubernetes,
and compare actual measured connectivity to expected connectivity using a truth table.

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

## Policy analysis

### Explain policies

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

### Which policy rules apply to a pod?

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


### Will policies allow or block traffic?

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

### Simulated probe

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

### Linter

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

## Developer guide

### Setup

 - [Get set up with golang 1.15](https://golang.org/dl/)
 - clone this repo

        git clone git@github.com:mattfenwick/cyclonus.git
        cd cyclonus

 - set up a KinD cluster with a CNI that supports network policies

        pushd kind/calico
        ./setup.sh
        popd

 - run cyclonus

        go run cmd/cyclonus/main.go generate --mode=example


## How to Release Binaries

See `goreleaser`'s requirements [here](https://goreleaser.com/environment/).

Get a [GitHub Personal Access Token](https://github.com/settings/tokens/new) and add the `repo` scope.
Set `GITHUB_TOKEN` to this value:

```bash
export GITHUB_TOKEN=...
```

[See here for more information on github tokens](https://help.github.com/articles/creating-an-access-token-for-command-line-use/).

Choose a tag/release name, create and push a tag:

```bash
TAG=v0.0.1

git tag $TAG
git push origin $TAG
```

Cut a release:

```bash
goreleaser release --rm-dist
```

Make a test release:

```bash
goreleaser release --snapshot --rm-dist
```

## Motivation and History

Testing network policys for CNI providers on Kubernetes has historically been very difficult, requiring alot of boiler plate.  This was recently improved upstream via truth table based tests (https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/1611-network-policy-validation).  Cyclonus is the next evolution of the truth table tests which are part of upstream Kubernetes.  Cyclonus generates 100s of truth tables, and outputs results in the same, easy to read format.
