# Cyclonus

## network policy explainer

Parse, explain, and probe network policies to understand their implications and help design
policies that suit your needs!

## Examples

### Probe

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

  {"Namespace": "default", "PodSelector": ["MatchLabels",["app: foo"],"MatchExpression",null]}
    source rules:
      default/allow-nothing
    all ingress blocked
  
  {"Namespace": "default", "PodSelector": ["MatchLabels",["app: bookstore","role: api"],"MatchExpression",null]}
    source rules:
      default/allow-from-app-bookstore-to-app-bookstore-role-api
    ingress:
      Internal:
        Namespace/Pod:
          namespace default
          pods matching ["MatchLabels",["app: bookstore"],"MatchExpression",null]
          Port(s):
            all ports all protocols
    
  {"Namespace": "default", "PodSelector": ["MatchLabels",null,"MatchExpression",null]}
    source rules:
      default/allow-no-egress-from-namespace
    all egress blocked
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