# Developer guide

## Setup

 - [Get set up with golang 1.16](https://golang.org/dl/)
 - clone this repo

        git clone git@github.com:mattfenwick/cyclonus.git
        cd cyclonus

 - set up a KinD cluster with a CNI that supports network policies

        pushd hack/kind/calico
          ./setup.sh
        popd

 - run cyclonus

        go run cmd/cyclonus/main.go generate --mock --perturbation-wait-seconds 0

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