# Cyclonus

## network policy explainer

1. read network policies

 - from a kubernetes cluster
 - from files
 
2. policy analysis

 - break the policies down by target, ingress/egress etc.

3. traffic analysis

 - given a pod (with labels, in a namespace with labels) determine which policies apply
 - given traffic between pods, determine whether it would be allowed or not
