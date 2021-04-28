# Cyclonus

## Network policy explainer, prober, and test case generator

Parse, explain, and probe network policies to understand their implications and help design
policies that suit your needs!

## Quickstart

Grab the [latest release](https://github.com/mattfenwick/cyclonus/releases) to get started using Cyclonus.

### Run as a kubernetes job

Create a job.yaml file:
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: cyclonus
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
        - command:
            - ./cyclonus
            - generate
          name: cyclonus
          imagePullPolicy: IfNotPresent
          image: mfenwick100/cyclonus:latest
      serviceAccount: cyclonus
```

Then create a namespace, service account, and job:
```bash
kubectl create ns netpol
kubectl create clusterrolebinding cyclonus --clusterrole=cluster-admin --serviceaccount=netpol:cyclonus
kubectl create sa cyclonus -n netpol
  
kubectl create -f job.yaml -n netpol
```

Use `kubectl logs -f` to watch your job go!

### Run on KinD

Take a look at the [kind directory](./hack/kind):

```
ls hack/kind 
antrea          calico          cilium          ovn-kubernetes  run-cyclonus.sh
```

Choose the right job for your CNI, then run:

```
cd hack/kind
CNI=calico RUN_FROM_SOURCE=false ./run-cyclonus.sh
```

This will:

 - create a namespace, service account, and cluster role binding for cyclonus
 - create a cyclonus job

Pull the logs from the job:

```
kubectl logs -f -n netpol cyclonus-abcde
```

### Run from source

Assuming you have a kube cluster and your kubectl is configured to point to it, you can run:

```
cd cmd/cyclonus
go run main.go generate
```

### Docker images

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

[Cyclonus runs network policy tests for Antrea on a daily basis](https://github.com/vmware-tanzu/antrea/actions/workflows/netpol_cyclonus.yml).

### Cilium testing

[Cyclonus runs network policy tests for Cilium on a daily basis](https://github.com/cilium/cilium/pull/14889).


## Cyclonus functionality

 - [run a single network policy test on a cluster](./docs/probe.md)
 - [run network policy conformance tests on a cluster](./docs/generator.md)
 - [analyze network policies](./docs/analyze.md).


## Sonobuoy plugin

Check out [our sonobuoy plugin](./hack/sonobuoy)!

## Developer guide

### Setup

 - [Get set up with golang 1.16](https://golang.org/dl/)
 - clone this repo

        git clone git@github.com:mattfenwick/cyclonus.git
        cd cyclonus

 - set up a KinD cluster with a CNI that supports network policies

        pushd kind/calico
        ./setup.sh
        popd

 - run cyclonus

        go run cmd/cyclonus/main.go generate --mode=example

 - run format, vet, tests

        make fmt
        make vet
        make test

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

Testing network policies for CNI providers on Kubernetes has historically been very difficult, requiring a lot of boiler plate.  This was recently improved upstream via truth table based tests ([see KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/1611-network-policy-validation)).  Cyclonus is the next evolution of the truth table tests which are part of upstream Kubernetes.  Cyclonus generates hundreds of network policies, their connectivity tables, and outputs results in the same, easy to read format.

## Thanks to contributors

 - @dougsland
 - @jayunit100
 - @johnSchnake
 - @enhaocui
