package registry

import (
	"fmt"
)

type EventType int

const (
	EventChildrenChange EventType = iota
)

type Event struct {
	Path   string
	Action EventType
}

func (e Event) String() string {
	return fmt.Sprintf("Event{Acotion{%d}, Path{%s}", e.Action, e.Path)
}
