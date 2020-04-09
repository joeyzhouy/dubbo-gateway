package zookeeper

import (
	"github.com/apache/dubbo-go/common/logger"
	"github.com/dubbogo/go-zookeeper/zk"
	"strings"
	"sync"
	"time"
)

type zkClient struct {
	Conn *zk.Conn
	sync.Mutex
	Timeout       time.Duration
	ZkAddresses   []string
	eventRegistry map[string][]*chan struct{}
}

func newZkClient(zkAddresses []string, timeout time.Duration) (*zkClient, error) {
	z := new(zkClient)
	conn, event, err := zk.Connect(zkAddresses, timeout)
	if err != nil {
		return nil, err
	}
	z.Conn = conn
	z.Timeout = timeout
	z.ZkAddresses = zkAddresses
	go z.operateZkEvent(event)
	return z, nil
}

func (z *zkClient) operateZkEvent(event <-chan zk.Event) {
LOOP:
	for {
		e := <-event
		switch e.State {
		case zk.StateDisconnected:
			logger.Warnf("zk{addr:%s} state is StateDisconnected, so close the zk client.", z.ZkAddresses)
			if z.Conn != nil {
				z.Conn.Close()
				z.Conn = nil
			}
			break LOOP
		case zk.StateConnected, zk.StateConnecting:

			if arr, ok := z.eventRegistry[e.Path]; ok && len(arr) > 0 {
				for _, c := range arr {
					*c <- struct{}{}
				}
			}
		default:
			switch e.Type {
			case zk.EventNodeChildrenChanged:
				logger.Infof("zkClient get zk node changed event{path:%s}", e.Path)
				z.Lock()
				for p, a := range z.eventRegistry {
					if strings.HasPrefix(p, e.Path) {
						for _, e := range a {
							*e <- struct{}{}
						}
					}
				}
				z.Unlock()
			}
		}
	}
}

func (z *zkClient) RegisterEvent(path string, event *chan struct{}) {
	if path == "" || event == nil {
		return
	}
	z.Lock()
	defer z.Unlock()
	arr := z.eventRegistry[path]
	arr = append(arr, event)

	z.eventRegistry[path] = arr
	logger.Debugf("zkClient register event{path:%s, ptr:%p}", path, event)
}

