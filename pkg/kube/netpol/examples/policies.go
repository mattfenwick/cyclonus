package examples

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
	"strings"
)

func label(key, val string) map[string]string {
	return map[string]string{key: val}
}

func LabelString(labels map[string]string) string {
	// 1. first, sort the keys so we get a deterministic answer
	var keys []string
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	// 2. now use the sorted keys to generate chunks
	var chunks []string
	for _, key := range keys {
		chunks = append(chunks, key, labels[key])
	}
	// 3. join
	return strings.Join(chunks, "-")
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/01-deny-all-traffic-to-an-application.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-deny-all
spec:
  podSelector:
    matchLabels:
      app: web
  ingress: []
*/
func AllowNothingTo(namespace string, toLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-nothing-to-%s", LabelString(toLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: toLabels},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// Same as above, but with empty slice instead of nil slice
func AllowNothingToEmptyIngress(namespace string, targetLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-nothing-to-v2-%s", LabelString(targetLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Ingress:     []networkingv1.NetworkPolicyIngressRule{},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/02-limit-traffic-to-an-application.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: api-allow
spec:
  podSelector:
    matchLabels:
      app: bookstore
      role: api
  ingress:
  - from:
      - podSelector:
          matchLabels:
            app: bookstore
*/
func AllowFromTo(namespace string, fromLabels map[string]string, toLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-from-%s-to-%s", LabelString(fromLabels), LabelString(toLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: toLabels},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{MatchLabels: fromLabels},
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/02a-allow-all-traffic-to-an-application.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-allow-all
  Namespace: default
spec:
  podSelector:
    matchLabels:
      app: web
  ingress:
  - {}
*/
func AllowAllTo(namespace string, toLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-all-to-%s", LabelString(toLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: toLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/03-deny-all-non-whitelisted-traffic-in-the-namespace.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: default-deny-all
  Namespace: default
spec:
  podSelector: {}
  ingress: []
*/
func AllowNothingToAnything(namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-nothing-to-anything",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress:     []networkingv1.NetworkPolicyIngressRule{},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/04-deny-traffic-from-other-namespaces.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  Namespace: secondary
  name: deny-from-other-namespaces
spec:
  podSelector:
    matchLabels:
  ingress:
  - from:
    - podSelector: {}
*/
func AllowAllWithinNamespace(namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-all-within-namespace",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/05-allow-traffic-from-all-namespaces.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  Namespace: secondary
  name: web-allow-all-namespaces
spec:
  podSelector:
    matchLabels:
      app: web
  ingress:
  - from:
    - namespaceSelector: {}
*/
func AllowAllTo_Version2(namespace string, targetLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-all-to-version2-%s", LabelString(targetLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{},
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  Namespace: secondary
  name: web-allow-all-namespaces
spec:
  podSelector:
    matchLabels:
      app: web
  ingress:
  - from:
*/
func AllowAllTo_Version3(namespace string, targetLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-all-to-version3-%s", LabelString(targetLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{From: nil},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

func AllowAllTo_Version4(namespace string, toLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-all-to-version4-%s", LabelString(toLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: toLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}, NamespaceSelector: &metav1.LabelSelector{}}},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/06-allow-traffic-from-a-namespace.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-allow-prod
spec:
  podSelector:
    matchLabels:
      app: web
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          purpose: production
*/
func AllowFromNamespaceTo(namespace string, namespaceLabels map[string]string, toLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-from-namespace-to-%s", LabelString(toLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: toLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{MatchLabels: namespaceLabels},
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/07-allow-traffic-from-some-pods-in-another-namespace.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-allow-all-ns-monitoring
  Namespace: default
spec:
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
            type: monitoring
*/
func AllowFromDifferentNamespaceWithLabelsTo(namespace string, fromLabels, namespaceLabels, toLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-from-namespace-with-labels-%s-to-%s", LabelString(fromLabels), LabelString(toLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: toLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector:       &metav1.LabelSelector{MatchLabels: fromLabels},
							NamespaceSelector: &metav1.LabelSelector{MatchLabels: namespaceLabels},
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/08-allow-external-traffic.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: web-allow-external
spec:
  podSelector:
    matchLabels:
      app: web
  ingress:
  - from: []
*/
func AllowFromAnywhere(namespace string, targetLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-from-anywhere-to-%s", LabelString(targetLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/09-allow-traffic-only-to-a-port.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: api-allow-5000
spec:
  podSelector:
    matchLabels:
      app: apiserver
  ingress:
  - ports:
    - port: 5000
    from:
    - podSelector:
        matchLabels:
          role: monitoring
*/
func AllowSpecificPortTo(namespace string, fromLabels, targetLabels map[string]string, targetPort int) *networkingv1.NetworkPolicy {
	portRef := intstr.FromInt(targetPort)
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-specific-port-from-%s-to-%s", LabelString(fromLabels), LabelString(targetLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Port: &portRef},
					},
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: fromLabels,
							},
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/10-allowing-traffic-with-multiple-selectors.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: redis-allow-services
spec:
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
          role: web
*/
func AllowFromMultipleTo(namespace string, fromLabels []map[string]string, targetLabels map[string]string) *networkingv1.NetworkPolicy {
	var froms []networkingv1.NetworkPolicyPeer
	for _, labels := range fromLabels {
		froms = append(froms, networkingv1.NetworkPolicyPeer{
			PodSelector: &metav1.LabelSelector{MatchLabels: labels},
		})
	}
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-from-multiple-to-%s", LabelString(targetLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{From: froms},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/11-deny-egress-traffic-from-an-application.md
/*
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
  egress: []
*/
func AllowNoEgressFromLabels(namespace string, targetLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-no-egress-from-labels-%s", LabelString(targetLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Egress:      []networkingv1.NetworkPolicyEgressRule{},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}
}

/*
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
      protocol: TCP
*/
func AllowEgressOnPort(namespace string, targetLabels map[string]string, port int) *networkingv1.NetworkPolicy {
	tcp := v1.ProtocolTCP
	udp := v1.ProtocolUDP
	portRef := intstr.FromInt(port)
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-egress-on-port-%s", LabelString(targetLabels)),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: &tcp, Port: &portRef},
						{Protocol: &udp, Port: &portRef},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/12-deny-all-non-whitelisted-traffic-from-the-namespace.md
/*
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: default-deny-all-egress
  Namespace: default
spec:
  policyTypes:
  - Egress
  podSelector: {}
  egress: []
*/
func AllowNoEgressFromNamespace(namespace string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-no-egress-from-namespace"),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Egress:      []networkingv1.NetworkPolicyEgressRule{},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}
}

// https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/14-deny-external-egress-traffic.md
/*
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
    to:
    - namespaceSelector: {}
*/
func AllowEgressToAllNamespacesOnPort(namespace string, targetLabels map[string]string, port int) *networkingv1.NetworkPolicy {
	tcp := v1.ProtocolTCP
	udp := v1.ProtocolUDP
	portRef := intstr.FromInt(port)
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-egress-to-all-namespace-from-%s-on-port-%d", LabelString(targetLabels), port),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: &tcp, Port: &portRef},
						{Protocol: &udp, Port: &portRef},
					},
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{},
						},
					},
				},
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}
}

/*
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-nothing
spec:
  podSelector:
    matchLabels:
      app: foo
  policyTypes:
  - Egress
  - Ingress
*/
func AllowNoIngressNorEgress(namespace string, targetLabels map[string]string) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("allow-nothing"),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
		},
	}
}

var AllExamples = []*networkingv1.NetworkPolicy{
	AllowNothingTo("default", map[string]string{"app": "web"}),
	AllowNothingToEmptyIngress("default", label("all", "web")),
	AllowFromTo("default", map[string]string{"app": "bookstore"}, map[string]string{"app": "bookstore", "role": "api"}),
	AllowAllTo("default", map[string]string{"app": "web"}),
	AllowNothingToAnything("default"),
	AllowAllWithinNamespace("default"),
	AccidentalAnd("default", label("a", "b"), label("user", "alice"), label("role", "client")),
	AccidentalOr("default", label("a", "b"), label("user", "alice"), label("role", "client")),
	AllowAllTo_Version2("default", label("app", "web")),
	AllowAllTo_Version3("default", label("app", "web")),
	AllowAllTo_Version4("default", label("app", "web")),
	AllowFromNamespaceTo("default", label("purpose", "production"), label("app", "web")),
	AllowFromDifferentNamespaceWithLabelsTo("default", label("type", "monitoring"), label("team", "operations"), label("app", "web")),
	AllowFromAnywhere("default", label("app", "web")),
	AllowSpecificPortTo("default", label("role", "monitoring"), label("app", "apiserver"), 5000),
	AllowFromMultipleTo(
		"default",
		[]map[string]string{
			{"app": "bookstore", "role": "search"},
			{"app": "bookstore", "role": "api"},
			{"app": "inventory", "role": "web"},
		},
		map[string]string{"app": "bookstore", "role": "db"}),
	AllowNoEgressFromLabels("default", label("app", "foo")),
	AllowEgressOnPort("default", label("app", "foo"), 53),
	AllowNoEgressFromNamespace("default"),
	AllowEgressToAllNamespacesOnPort("default", label("app", "foo"), 53),
	AllowNoIngressNorEgress("default", label("app", "foo")),
}
