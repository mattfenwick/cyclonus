# Cyclonus

## network policy explainer, prober, and fuzzer!

Parse, explain, and probe network policies to understand their implications and help design
policies that suit your needs!

Grab the [latest release](https://github.com/mattfenwick/cyclonus/releases) to get started using Cyclonus!

## Examples

### Probe

Run a connectivity probe against a Kubernetes cluster.


```
$ go run cmd/cyclonus/main.go probe --noisy=true

INFO[2021-01-20T06:30:25-05:00] found 1 policies across namespaces [x y z]   
{"Namespace": "x", "PodSelector": ["MatchLabels",["pod: a"],"MatchExpression",null]}
  source rules:
    x/vary-ingress-empty
  all ingress blocked


+---------+------------------+---------------------+-----------------+------------------------+----------------------+
|  TYPE   | TARGET NAMESPACE | TARGET POD SELECTOR |      PEER       |     PORT/PROTOCOL      |     SOURCE RULES     |
+---------+------------------+---------------------+-----------------+------------------------+----------------------+
| Ingress | x                | Match labels:       |                 |                        | x/vary-ingress-empty |
|         |                  |   pod: a            |                 |                        |                      |
+---------+------------------+---------------------+-----------------+------------------------+----------------------+
|         |                  |                     | no pods, no ips | no ports, no protocols |                      |
+---------+------------------+---------------------+-----------------+------------------------+----------------------+
INFO[2021-01-20T06:30:25-05:00] synthetic probe on port 80, protocol TCP     
INFO[2021-01-20T06:30:25-05:00] running probe on port 80, protocol TCP       
INFO[2021-01-20T06:30:25-05:00] kube probe on port 80, protocol TCP          


Kube results:
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
found 81 true, 0 false, 0 no value from 81 total
```

### Fuzzer

Generate network policies, install the policies one at a time in kubernetes, and compare actual measured connectivity
to expected connectivity using a truth table.

```
$ go run cmd/cyclonus/main.go fuzz \
  --mode vary-ingress \
  --noisy=true \
  --netpol-creation-wait-seconds 15

... (snip) ...

Kube results for x/vary-ingress-28-0-0-0-10:
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
Discrepancy found: 9 wrong, 72 no value, 0 correct out of 81 total
Ingress:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|  -  | X/A | X/B | X/C | Y/A | Y/B | Y/C | Z/A | Z/B | Z/C |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
| x/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| x/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| x/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
Egress:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|  -  | X/A | X/B | X/C | Y/A | Y/B | Y/C | Z/A | Z/B | Z/C |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
| x/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| x/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| x/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
Combined:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|  -  | X/A | X/B | X/C | Y/A | Y/B | Y/C | Z/A | Z/B | Z/C |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
| x/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| x/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| x/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| y/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/a | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/b | .   | .   | .   | .   | .   | .   | .   | .   | .   |
| z/c | .   | .   | .   | .   | .   | .   | .   | .   | .   |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+

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

... (snip) ...

```

### Synthetic Probe

Using hypothetical "traffic", generate a table of presumed connectivity by evaluating network
policies.  Note: this does not use a kubernetes cluster.

```
$ go run cmd/cyclonus/main.go probe \
  --model-path cmd/cyclonus/probe-example.json \
  --policy-source examples

Ingress:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | X/C | X/A | X/B | DEFAULT/A | DEFAULT/B | DEFAULT/C | Z/C | Z/A | Z/B |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| x/c       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| x/a       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| x/b       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| default/a | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/b | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/c | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| z/c       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| z/a       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| z/b       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
Egress:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | X/C | X/A | X/B | DEFAULT/A | DEFAULT/B | DEFAULT/C | Z/C | Z/A | Z/B |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| x/c       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| x/a       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| x/b       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/a | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| default/b | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| default/c | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| z/c       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| z/a       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| z/b       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
Combined:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | X/C | X/A | X/B | DEFAULT/A | DEFAULT/B | DEFAULT/C | Z/C | Z/A | Z/B |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| x/c       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| x/a       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| x/b       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| default/a | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| default/b | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| default/c | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| z/c       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| z/a       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| z/b       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
```

## Analyze

Groups policies by target, divides rules into egress and ingress, and gives a basic explanation of the combined
policies.  This clarifies the interactions between "denies" and "allows" from multiple policies.

```
$ go run cmd/cyclonus/main.go analyze \
  --policy-source examples 

+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
|  TYPE   | TARGET NAMESPACE | TARGET POD SELECTOR |           PEER           |       PORT/PROTOCOL       |            SOURCE RULES              |
+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
| Ingress | default          | Match labels:       |                          |                           | default/accidental-and               |
|         |                  |   a: b              |                          |                           | default/accidental-or                |
+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
|         |                  |                     | no ips                   | no ports, no protocols    |                                      |
+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
|         |                  |                     | namespace: Match labels: | all ports, all protocols  |                                      |
|         |                  |                     |   user: alice            |                           |                                      |
|         |                  |                     | pods: Match labels:      |                           |                                      |
|         |                  |                     |   role: client           |                           |                                      |
+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
|         |                  |                     | namespace: Match labels: | all ports, all protocols  |                                      |
|         |                  |                     |   user: alice            |                           |                                      |
|         |                  |                     | pods: all                |                           |                                      |
+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
|         |                  |                     | namespace: default       | all ports, all protocols  |                                      |
|         |                  |                     | pods: Match labels:      |                           |                                      |
|         |                  |                     |   role: client           |                           |                                      |
+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
| Ingress | default          | Match labels:       |                          |                           | default/allow-nothing-to-v2-all-web  |
|         |                  |   all: web          |                          |                           |                                      |
+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
|         |                  |                     | no pods, no ips          | no ports, no protocols    |                                      |
+---------+------------------+---------------------+--------------------------+---------------------------+--------------------------------------+
```

## traffic

Given arbitrary traffic examples (from a source to a destination, including labels, over a port and protocol),
this command parses network policies and determines if the traffic is allowed or not.

```
$ go run cmd/cyclonus/main.go traffic \
  --policy-source examples \
  --traffic-path cmd/cyclonus/traffic-example.json 

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

TODO in progress


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
