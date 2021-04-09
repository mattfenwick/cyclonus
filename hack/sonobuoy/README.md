# Sonobuoy plugin

## Create plugin

```bash
sonobuoy gen plugin \
  --name=cyclonus \
  --image=docker.io/mfenwick100/cyclonus:latest \
  --cmd cyclonus \
  --cmd generate \
  --cmd="--include=conflict" > cyclonus-plugin.yaml
```

## Run plugin

```bash
sonobuoy run --plugin cyclonus-plugin.yaml --wait
```

## TODO

May need an 'updateProgress' function, see
https://github.com/vmware-tanzu/sonobuoy/blob/v0.50.0/examples/plugins/progress-reporter/run.sh#L23.

## Look at results

TODO, maybe:
```bash
outfile=$(sonobuoy retrieve) && \
  mkdir results && tar -xf $outfile -C results &&
  cat results/plugins/debug/results/*/out*
```