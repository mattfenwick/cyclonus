# Cyclonus

## network policy explainer, prober, and test case generator!

Parse, explain, and probe network policies to understand their implications and help design
policies that suit your needs!

Grab the [latest release](https://github.com/mattfenwick/cyclonus/releases) to get started using Cyclonus!

## Examples

### Probe

Run a connectivity probe against a Kubernetes cluster.

```
$ go run cmd/cyclonus/main.go probe

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

0 wrong, 81 no value, 0 correct, 0 ignored out of 81 total
```

### Generator

Generate network policies, install the policies one at a time in kubernetes, and compare actual measured connectivity
to expected connectivity using a truth table.

```
$ go run cmd/cyclonus/main.go generate \
  --mode simple-fragments \
  --netpol-creation-wait-seconds 15

... 
Synthetic vs combined:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|  -  | X/A | X/B | X/C | Y/A | Y/B | Y/C | Z/A | Z/B | Z/C |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
| x/a | X   | .   | .   | .   | .   | .   | .   | .   | .   |
| x/b | X   | .   | .   | .   | .   | .   | .   | .   | .   |
| x/c | X   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/a | X   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/b | X   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/c | X   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/a | X   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/b | X   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/c | X   | .   | .   | .   | .   | .   | .   | .   | .   |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
... 
```

### Synthetic Probe

Using hypothetical "traffic", generate a table of presumed connectivity by evaluating network
policies.  Note: this does not use a kubernetes cluster.

```
$ go run cmd/cyclonus/main.go analyze \
  --probe-path examples/synthetic-probe-example.json \
  --policy-source examples
  
Combined:
+-----------+-----------+-----------+-----------+-----+-----+-----+-----+-----+-----+
|     -     | DEFAULT/A | DEFAULT/B | DEFAULT/C | X/A | X/B | X/C | Y/A | Y/B | Y/C |
+-----------+-----------+-----------+-----------+-----+-----+-----+-----+-----+-----+
| default/a | X         | X         | X         | X   | X   | X   | X   | X   | X   |
| default/b | X         | X         | X         | X   | X   | X   | X   | X   | X   |
| default/c | X         | X         | X         | X   | X   | X   | X   | X   | X   |
| x/a       | X         | X         | X         | .   | .   | .   | .   | .   | .   |
| x/b       | X         | X         | X         | .   | .   | .   | .   | .   | .   |
| x/c       | X         | X         | X         | .   | .   | .   | .   | .   | .   |
| y/a       | X         | X         | X         | .   | .   | .   | .   | .   | .   |
| y/b       | X         | X         | X         | .   | .   | .   | .   | .   | .   |
| y/c       | X         | X         | X         | .   | .   | .   | .   | .   | .   |
+-----------+-----------+-----------+-----------+-----+-----+-----+-----+-----+-----+
```

## Explain

Groups policies by target, divides rules into egress and ingress, and gives a basic explanation of the combined
policies.  This clarifies the interactions between "denies" and "allows" from multiple policies.

```
$ go run cmd/cyclonus/main.go analyze \
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

## traffic

Given arbitrary traffic examples (from a source to a destination, including labels, over a port and protocol),
this command parses network policies and determines if the traffic is allowed or not.

```
$ go run cmd/cyclonus/main.go analyze \
  --policy-source examples \
  --traffic-path ./examples/traffic-example.json

Traffic:
{
  "Source": {
    "Internal": {
      "PodLabels": {
        "pod": "a"
      },
      "NamespaceLabels": {
        "ns": "z"
      },
      "Namespace": "z"
    },
    "IP": "192.168.1.13"
  },
  "Destination": {
    "Internal": {
      "PodLabels": {
        "pod": "b"
      },
      "NamespaceLabels": {
        "ns": "x"
      },
      "Namespace": "x"
    },
    "IP": "192.168.1.14"
  },
  "PortProtocol": {
    "Protocol": "TCP",
    "Port": 80
  }
}

Is allowed: true
```

## targets

Given a set of pods, this command determines which network policies affect those pods.

```
➜  cyclonus git:(master) ✗ go run ./cmd/cyclonus/main.go analyze \
  --explain=false \
  --policy-path ./networkpolicies/simple-example \
  --target-pod-path ./examples/targets-example.json

Combined:
+---------+---------------+-----------------------+-------------------+-------------------------+
|  TYPE   |    TARGET     |     SOURCE RULES      |       PEER        |      PORT/PROTOCOL      |
+---------+---------------+-----------------------+-------------------+-------------------------+
| Ingress | namespace: y  | y/allow-all-for-label | ports for all IPs | port 53 on protocol TCP |
|         | Match labels: | y/deny-all            |                   |                         |
|         |   pod: b      |                       |                   |                         |
+         +               +                       +-------------------+                         +
|         |               |                       | namespace: all    |                         |
|         |               |                       | pods: all         |                         |
|         |               |                       |                   |                         |
+---------+---------------+-----------------------+-------------------+-------------------------+

Matching targets:
+---------+---------------+-----------------------+-------------------+-------------------------+
|  TYPE   |    TARGET     |     SOURCE RULES      |       PEER        |      PORT/PROTOCOL      |
+---------+---------------+-----------------------+-------------------+-------------------------+
| Ingress | namespace: y  | y/allow-all-for-label | ports for all IPs | port 53 on protocol TCP |
|         | Match labels: |                       |                   |                         |
|         |   pod: b      |                       |                   |                         |
+         +               +                       +-------------------+                         +
|         |               |                       | namespace: all    |                         |
|         |               |                       | pods: all         |                         |
|         |               |                       |                   |                         |
+         +---------------+-----------------------+-------------------+-------------------------+
|         | namespace: y  | y/deny-all            | no pods, no ips   | no ports, no protocols  |
|         | all pods      |                       |                   |                         |
+---------+---------------+-----------------------+-------------------+-------------------------+
```


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
