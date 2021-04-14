# Sonobuoy plugin

## Create plugin

```bash
sonobuoy gen plugin \
  --name=cyclonus \
  --image=mfenwick100/cyclonus:latest \
  --cmd ./run-sonobuoy-plugin.sh \ > cyclonus-plugin.yaml
```

## Run plugin

```bash
sonobuoy run --plugin cyclonus-plugin.yaml --wait
```

## Look at results

```bash
outfile=$(sonobuoy retrieve) && \
  mkdir results && tar -xf $outfile -C results
```

Then crack open the `results` dir and have a look!
