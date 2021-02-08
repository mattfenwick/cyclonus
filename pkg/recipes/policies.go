package recipes

import (
	. "github.com/mattfenwick/cyclonus/pkg/connectivity/types"
	v1 "k8s.io/api/core/v1"
)

var container = []*Container{{Port: 80, Protocol: v1.ProtocolTCP}}
var cont5000 = []*Container{{Port: 5000, Protocol: v1.ProtocolTCP}}

const Recipe01 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-deny-all
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: web
  ingress: []`

var Resources01 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", map[string]string{"app": "web"}, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe02 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: api-allow
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: bookstore
      role: api
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: bookstore`

var Resources02 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", map[string]string{"app": "bookstore"}, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", map[string]string{"app": "bookstore"}, "", container),
		NewPod("default", "b", map[string]string{"app": "bookstore", "role": "api"}, "", container),
		NewPod("default", "c", map[string]string{"role": "api"}, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", map[string]string{"app": "bookstore"}, "", container),
	},
}

// TODO clarify that this allows ingress from outside the cluster
const Recipe02A = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-allow-all
  namespace: default
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: web
  ingress:
    - {}`

var Resources02A = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", map[string]string{"app": "web"}, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe03 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: default-deny-all
  namespace: default
spec:
  policyTypes:
    - Ingress
  podSelector: {}
  ingress: []`

var Resources03 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", nil, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe04 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  namespace: secondary
  name: deny-from-other-namespaces
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
  ingress:
    - from:
        - podSelector: {}`

var Resources04 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":         {},
		"default":   {},
		"secondary": {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", nil, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("secondary", "a", nil, "", container),
		NewPod("secondary", "b", nil, "", container),
		NewPod("secondary", "c", nil, "", container),
	},
}

const Recipe05 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  namespace: default
  name: web-allow-all-namespaces
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: web
  ingress:
    - from:
        - namespaceSelector: {}`

var Resources05 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", map[string]string{"app": "web"}, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe06 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-allow-prod
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: web
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              purpose: production`

var Resources06 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {"purpose": "production"},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", map[string]string{"app": "web"}, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe07 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-allow-all-ns-monitoring
  namespace: default
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: web
  ingress:
    - from:
        - namespaceSelector:     # chooses all pods in namespaces labelled with team=operations
            matchLabels:
              team: operations
          podSelector:           # chooses pods with type=monitoring
            matchLabels:
              type: monitoring`

var Resources07 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {"team": "operations"},
		"default": {},
		"y":       {"team": "operations"},
	},
	Pods: []*Pod{
		NewPod("x", "a", map[string]string{"type": "monitoring"}, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", map[string]string{"type": "monitoring"}, "", container),
		NewPod("default", "b", map[string]string{"app": "web"}, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", map[string]string{"type": "monitoring"}, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

// TODO clarify that this allows ingress from outside the cluster
const Recipe08 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-allow-external
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: web
  ingress:
    - from: []`

var Resources08 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", map[string]string{"app": "web"}, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe09 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: api-allow-5000
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: apiserver
  ingress:
    - ports:
        - port: 5000
      from:
        - podSelector:
            matchLabels:
              role: monitoring`

// TODO this example could be improved by adding some probes over a different port
var Resources09 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", map[string]string{"role": "monitoring"}, "", cont5000),
		NewPod("x", "b", nil, "", cont5000),
		NewPod("x", "c", nil, "", cont5000),
		NewPod("default", "a", map[string]string{"role": "monitoring"}, "", cont5000),
		NewPod("default", "b", map[string]string{"app": "apiserver"}, "", cont5000),
		NewPod("default", "c", nil, "", cont5000),
		NewPod("y", "a", map[string]string{"role": "monitoring"}, "", cont5000),
		NewPod("y", "b", nil, "", cont5000),
		NewPod("y", "c", nil, "", cont5000),
	},
}

const Recipe10 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: redis-allow-services
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: bookstore
      role: db
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: bookstore
              role: search
        - podSelector:
            matchLabels:
              app: bookstore
              role: api
        - podSelector:
            matchLabels:
              app: inventory
              role: web`

var Resources10 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", map[string]string{"app": "bookstore", "role": "search"}, "", container),
		NewPod("default", "b", map[string]string{"app": "bookstore", "role": "db"}, "", container),
		NewPod("default", "c", map[string]string{"app": "bookstore", "role": "api"}, "", container),
		NewPod("default", "d", map[string]string{"app": "inventory", "role": "web"}, "", container),
		NewPod("y", "a", map[string]string{"app": "bookstore", "role": "search"}, "", container),
		NewPod("y", "b", map[string]string{"app": "bookstore", "role": "api"}, "", container),
		NewPod("y", "c", map[string]string{"app": "inventory", "role": "web"}, "", container),
	},
}

const Recipe11_1 = `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: foo-deny-egress
spec:
  podSelector:
    matchLabels:
      app: foo
  policyTypes:
    - Egress
  egress: []`

var Resources11_1 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", map[string]string{"app": "foo"}, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe11_2 = `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: foo-deny-egress
spec:
  podSelector:
    matchLabels:
      app: foo
  policyTypes:
    - Egress
  egress:
    # allow DNS resolution
    - ports:
        - port: 53
          protocol: UDP
        - port: 53
          protocol: TCP`

var Resources11_2 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", map[string]string{"app": "foo"}, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe12 = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: default-deny-all-egress
  namespace: default
spec:
  policyTypes:
    - Egress
  podSelector: {}
  egress: []`

var Resources12 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", nil, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}

const Recipe14 = `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: foo-deny-external-egress
spec:
  podSelector:
    matchLabels:
      app: foo
  policyTypes:
    - Egress
  egress:
    - ports:
        - port: 53
          protocol: UDP
        - port: 53
          protocol: TCP
    - to:
        - namespaceSelector: {}`

var Resources14 = &Resources{
	Namespaces: map[string]map[string]string{
		"x":       {},
		"default": {},
		"y":       {},
	},
	Pods: []*Pod{
		NewPod("x", "a", nil, "", container),
		NewPod("x", "b", nil, "", container),
		NewPod("x", "c", nil, "", container),
		NewPod("default", "a", nil, "", container),
		NewPod("default", "b", nil, "", container),
		NewPod("default", "c", nil, "", container),
		NewPod("y", "a", nil, "", container),
		NewPod("y", "b", nil, "", container),
		NewPod("y", "c", nil, "", container),
	},
}
