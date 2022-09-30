# Cyclonus

## Network policy explainer, prober, and test case generator

Parse, explain, and probe network policies to understand their implications and help design
policies that suit your needs!

## Quickstart

Users: check out our [Quickstart guide](./docs/quickstart.md)

Developers: check out our [Developer guide](./docs/developer-guide.md)

Cyclonus functionality:

 - [run a single network policy test on a cluster](./docs/probe.md)
 - [run network policy conformance tests on a cluster](./docs/generator.md)
 - [understand test runs](./docs/test-runs.md)
 - [analyze network policies](./docs/analyze.md)


## Cyclonus disambiguation

If you're looking for the official kubernetes network policy tool, check out: https://github.com/kubernetes-sigs/network-policy-api .  Actively development is continuing there.

https://github.com/mattfenwick/cyclonus will remain open and be used for research, experimentation, and hacking!

## Integrations

Cyclonus is available as a [**krew/kubectl plugin**](https://github.com/mattfenwick/kubectl-cyclonus):

 - [Set up krew](https://krew.sigs.k8s.io/docs/user-guide/quickstart/)
 - install: `kubectl krew install cyclonus`
 - use: `kubectl cyclonus -h`

**Antrea testing**: [Cyclonus runs network policy tests for Antrea on a daily basis](https://github.com/vmware-tanzu/antrea/actions/workflows/netpol_cyclonus.yml).

**Cilium testing**: [Cyclonus runs network policy tests for Cilium on a daily basis](https://github.com/cilium/cilium/pull/14889).

**Sonobuoy plugin**: [run Cyclonus tests through Sonobuoy](./hack/sonobuoy).


## Motivation and History

Testing network policies for CNI providers on Kubernetes has historically been very difficult, requiring a lot of boiler plate.
This was recently improved upstream via truth table based tests 
([see KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/1611-network-policy-validation)).
Cyclonus is the next evolution of the truth table tests which are part of upstream Kubernetes.
Cyclonus generates hundreds of network policies, their connectivity tables, and outputs results in the same, easy to read format.

## Thanks to contributors

 - @dougsland
 - @jayunit100
 - @johnSchnake
 - @enhaocui
 - @matmerr
