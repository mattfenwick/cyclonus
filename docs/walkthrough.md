# Cyclonus: kubernetes network policy spec and CNI conformance

Cyclonus is a kubernetes network policy tool which can:

 - show you what, precisely, your network policies mean according to the kubernetes spec
 - verify that your CNI conforms to the kubernetes network policy standard

# Why Cyclonus?

See [Cyclonus motivation and history](https://github.com/mattfenwick/cyclonus#motivation-and-history).

Example of subtle/confusing policies: 

 - https://kubernetes.io/docs/concepts/services-networking/network-policies/#behavior-of-to-and-from-selectors
 - [isolation semantics](https://kubernetes.io/docs/concepts/services-networking/network-policies/#the-two-sorts-of-pod-isolation)
 - [inability to directly block traffic](https://kubernetes.io/docs/concepts/services-networking/network-policies/#what-you-can-t-do-with-network-policies-at-least-not-yet)

# Cyclonus secret sauce

Visualization of a network policy's meaning -- TODO explain

Network policy engine: executable network policy specification -- TODO explain

# Code walkthrough

## Network policy engine

[`package matcher`](../pkg/matcher)

Goals:
 - models:
   - network policy
   - traffic: request from source->destination, using a port and protocol
 - leverage models:
   - ingest k8s network policy
   - given some network policies, will traffic be blocked or allowed? 
   - generate human-readable explanation of network policies

Key types:
 - Policy
   - BuildNetworkPolicies
   - Policy.Simplify
   - Policy.IsTrafficAllowed
   - Policy.ExplainTable
 - Target
   - BuildTarget
 - Traffic

## Kubernetes wrapper and mock

[`package kube`](../pkg/kube)

Goals: leverage kubernetes API server to:
 - update state of cluster
 - run commands (similar to `kubectl exec`)

key types:
 - Kubernetes
   - ExecuteRemoteCommand
 - MockKubernetes

## Network policy test scenario generator

[`package generator`](../pkg/generator)

Goals:
 - model of network policy test scenario -- data collection, cluster perturbation
 - generate network policy test scenarios, whic explore various features of network 
   policies including subtle areas of syntax and semantics
 - organize test scenarios by features used, to enable focus in testing and analysis of results

Key types:
 - Tag (string)
 - TestCase
 - TestCaseGenerator

## Connectivity test

[`package connectivity`](../pkg/connectivity)

Goals:
 - execute test case on a kubernetes cluster
 - meaningfully compare actual to expected results

How is a test case executed?
 1. blank slate; collect connectivity table
 2. for each (1+) perturbation:
   - update cluster state
   - collect connectivity data

Key types:
 - ComparisonTable
 - Interpreter
   - Interpreter.ExecuteTestCase
 - Result / StepResult
 - Printer

### Connectivity test: data collection

[`package probe`](../pkg/connectivity/probe)

Goals:
 - describe desired state of cluster before collection of connectivity data
 - use kube package to collect data for connectivity table

Key types:
 - Connectivity
 - Job
   - Job.KubeExecCommand
 - Resources
   - Resources.CreateResourcesInKube
 - Runner / KubeJobRunner / SimulatedJobRunner
   - RunJobs
 - TruthTable


# Using Cyclonus

[CLI usage](../README.md#cli-usage)

Examples:
 - [`cyclonus analyze`](../examples/run.sh)
 - `cyclonus probe`:

   ```bash
   go run cmd/cyclonus/main.go probe \
     --server-protocol=tcp \
     --server-port=80
   ```

 - `cyclonus generate`:
   
   ```bash
   go run cmd/cyclonus/main.go generate \
     --include conflict \
     --job-timeout-seconds 2 \
     --ignore-loopback=true \
     --server-protocol=tcp \
     --server-port=80 \
     --mock \
     --perturbation-wait-seconds=0
   ```

# Next steps

Work with community to build out Cyclonus functionality and evolve in tandem with kubernetes network policy:
https://github.com/kubernetes-sigs/network-policy-api/tree/master/cmd/cyclonus .

# Acknowledgements

Collaborators:
 - Jay
 - Amim
 - Ricardo

#sig-network

CNI teams:
 - Antrea
 - ovn-kubernetes
 - Cilium
 - Calico
