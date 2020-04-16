package common

import "fmt"

type Node struct {
	IP   string
	Port int
}

func (n *Node) String() string {
	return fmt.Sprintf("%s:%d", n.IP, n.Port)
}
