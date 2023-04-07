# Developer guide

## Setup

 - [Get set up with golang 1.20](https://golang.org/dl/)
 - clone this repo

        git clone git@github.com:mattfenwick/cyclonus.git
        cd cyclonus

 - set up a KinD cluster with a CNI that supports network policies

        pushd hack/kind/calico
          ./setup-kind.sh
        popd

 - run cyclonus

        go run cmd/cyclonus/main.go generate --mock --perturbation-wait-seconds 0

 - run format, vet, tests

        make fmt
        make vet
        make test

## How to upgrade k8s library version

1. go to [a kubernetes repo](https://github.com/kubernetes/apimachinery/tags)
2. look at the tag versions
3. choose the latest release version
4. update the k8s.io library versions in [go.mod](../go.mod)

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

You'll need to be logged in to dockerhub from the command line, in order to push images: `docker login`

Cut a release:

```bash
goreleaser release --rm-dist
```

Make a test release:

```bash
goreleaser release --snapshot --rm-dist
```
