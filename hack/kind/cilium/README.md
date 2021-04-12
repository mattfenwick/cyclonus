## Running Cilium on Fedora 33

coredns, local path storage containers get stuck in CreatingContainer status forever with no endpoint available.

Workaround: change the helm invocation to (this sets `hostServices.enabled=true` instead of `false`):

```bash
helm install cilium cilium/cilium \
  --version "${CILIUM_VERSION}" \
  --namespace kube-system \
  --set nodeinit.enabled=true \
  --set kubeProxyReplacement=partial \
  --set hostServices.enabled=true \
  --set externalIPs.enabled=true \
  --set nodePort.enabled=true \
  --set hostPort.enabled=true \
  --set bpf.masquerade=false \
  --set image.pullPolicy=IfNotPresent \
  --set ipam.mode=kubernetes
```

NOTE: The issue might affect others distros with kernel > 5.10

Tested with: Fedora 33, Kernel: 5.10.19-200.fc33.x86_64

See also:

- https://github.com/cilium/cilium/issues/14960
- https://github.com/cilium/cilium/pull/14951