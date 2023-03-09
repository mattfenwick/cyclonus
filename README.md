# Cyclonus

## Network policy explainer, prober, and test case generator

Parse, explain, and probe network policies to understand their implications and help design
policies that suit your needs!

## Quickstart

Users: check out the:

 - [Quickstart guide](./docs/quickstart.md)
 - [understand test runs](./docs/test-runs.md)

Developers: check out the [Developer guide](./docs/developer-guide.md)

### CLI usage

 - `cyclonus analyze`: [leverage network policy engine to precisely understand your policies](./docs/command-analyze.md)
 - `cyclonus generate`: [run network policy conformance test suites on a cluster](./docs/command-generate.md)
 - `cyclonus probe`: [run a single network policy test on a cluster](./docs/command-probe.md)


## Cyclonus disambiguation

If you're looking for the official kubernetes network policy tool, check out: https://github.com/kubernetes-sigs/network-policy-api .  Active development of network policy verification is continuing there.

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

Testing network policies for CNI providers on Kubernetes has historically been very difficult,
requiring a lot of boilerplate.

This was recently improved upstream via truth table based tests:

 - KEP: ["Architecting NetworkPolicy tests with a DSL for better upstream test coverage of all CNIs"](https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/1611-network-policy-validation)
 - kubernetes PR: ["new NetworkPolicy Validation suite"](https://github.com/kubernetes/kubernetes/pull/91592)

Cyclonus is the next evolution: in addition to truth-table connectivity tests, it adds two new components:
 - a powerful network policy engine implementing the Kubernetes network policy specification
 - a test case generator, allowing for easy testing of hundreds of network policy scenarios

Cyclonus aims to make network policies and implementations easy to understand, easy to use correctly, and easy to verify.

## Thanks to contributors

 - @dougsland
 - @jayunit100
 - @johnSchnake
 - @enhaocui
 - @matmerr
 - @huntergregory
