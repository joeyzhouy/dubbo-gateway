package registry

import (
	"dubbo-gateway/common/extension"
	"fmt"
)

type EventType int

const (
	EventTypeAdd = iota
	EventTypeDel
	EventTypeUpdate
)

type Event struct {
	Node   extension.Node
	Action EventType
}

func (e Event) String() string {
	return fmt.Sprintf("Event{Acotion{%d}, Path{%s}}", e.Action, e.Node)
}
