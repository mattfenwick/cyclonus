package probe

type Node struct {
	Name   string
	Labels map[string]string
	IP     string
}

func NewNode(name string, labels map[string]string, ip string) *Node {
	return &Node{
		Name:   name,
		Labels: make(map[string]string),
		IP:     ip,
	}
}
