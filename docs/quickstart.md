# Quickstart guide

Grab the [latest release](https://github.com/mattfenwick/cyclonus/releases) to get started using Cyclonus.

Images are available at [mfenwick100/cyclonus](https://hub.docker.com/r/mfenwick100/cyclonus/tags?page=1&ordering=last_updated):

```
docker pull docker.io/mfenwick100/cyclonus:latest
```

## Run as a kubernetes job

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

## Run on KinD

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

## Run from source

Assuming you have a kube cluster and your kubectl is configured to point to it, you can run:

```
go run cmd/cyclonus/main.go generate
```
