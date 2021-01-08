# Cyclonus

## network policy explainer

1. read network policies

 - from a kubernetes cluster
 - from files
 
2. policy analysis

 - break the policies down by target, ingress/egress etc.

3. traffic analysis

 - given a pod (with labels, in a namespace with labels) determine which policies apply
 - given traffic between pods, determine whether it would be allowed or not

4. connectivity probe

## Examples

### Connectivity probe

Using policies from files:
```
$ go run cmd/netpol-explainer/main.go probe --policy-source file --policy-path ./networkpolicies --model-namespace m,default,n

Ingress:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | M/A | M/B | M/C | DEFAULT/A | DEFAULT/B | DEFAULT/C | N/A | N/B | N/C |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| m/a       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| m/b       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| m/c       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| default/a | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| default/b | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| default/c | .   | .   | .   | .         | .         | X         | .   | .   | .   |
| n/a       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| n/b       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| n/c       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
Egress:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | M/A | M/B | M/C | DEFAULT/A | DEFAULT/B | DEFAULT/C | N/A | N/B | N/C |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| m/a       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| m/b       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| m/c       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/a | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/b | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/c | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| n/a       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| n/b       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| n/c       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
Combined:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | M/A | M/B | M/C | DEFAULT/A | DEFAULT/B | DEFAULT/C | N/A | N/B | N/C |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| m/a       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| m/b       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| m/c       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| default/a | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| default/b | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| default/c | .   | .   | .   | .         | .         | X         | .   | .   | .   |
| n/a       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| n/b       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
| n/c       | .   | .   | .   | X         | .         | X         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
```

Using policies from golang structs:

```
$ go run cmd/netpol-explainer/main.go probe --policy-source examples --model-namespace m,default,n

Ingress:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | M/A | M/B | M/C | DEFAULT/A | DEFAULT/B | DEFAULT/C | N/A | N/B | N/C |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| m/a       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| m/b       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| m/c       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| default/a | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/b | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/c | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| n/a       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| n/b       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| n/c       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
Egress:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | M/A | M/B | M/C | DEFAULT/A | DEFAULT/B | DEFAULT/C | N/A | N/B | N/C |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| m/a       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| m/b       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| m/c       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| default/a | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| default/b | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| default/c | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| n/a       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| n/b       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
| n/c       | .   | .   | .   | .         | .         | .         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
Combined:
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
|     -     | M/A | M/B | M/C | DEFAULT/A | DEFAULT/B | DEFAULT/C | N/A | N/B | N/C |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
| m/a       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| m/b       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| m/c       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| default/a | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| default/b | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| default/c | X   | X   | X   | X         | X         | X         | X   | X   | X   |
| n/a       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| n/b       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
| n/c       | .   | .   | .   | X         | X         | X         | .   | .   | .   |
+-----------+-----+-----+-----+-----------+-----------+-----------+-----+-----+-----+
```

## Analyze policies

```
$ go run cmd/netpol-explainer/main.go analyze --policy-source file --policy-path ./networkpolicies

{"Namespace": "default", "PodSelector": ["MatchLabels",null,"MatchExpression",null]}
  source rules:
    default/deny-all
  all ingress blocked

{"Namespace": "default", "PodSelector": ["MatchLabels",["pod: b"],"MatchExpression",null]}
  source rules:
    default/allow-all-for-label
  ingress:
  - anywhere: all pods in all namespaces and all IPs
    all ports all protocols

{"Namespace": "default", "PodSelector": ["MatchLabels",["pod: a"],"MatchExpression",null]}
  source rules:
    default/allow-label-to-label
    default/deny-all-for-label
  ingress:
  - pods matching ["MatchLabels",["pod: c"],"MatchExpression",null] in namespace default
    all ports all protocols
```
