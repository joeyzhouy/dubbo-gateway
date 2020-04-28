package zookeeper

import (
	"github.com/apache/dubbo-go/common/logger"
	"github.com/dubbogo/go-zookeeper/zk"
	perrors "github.com/pkg/errors"
	"path"
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
	z, event, err := newZkClientWithEvent(zkAddresses, timeout)
	if err != nil {
		return nil, err
	}
	go z.operateZkEvent(event)
	return z, nil
}

func newZkClientWithEvent(zkAddresses []string, timeout time.Duration) (*zkClient, <-chan zk.Event, error) {
	z := new(zkClient)
	conn, event, err := zk.Connect(zkAddresses, timeout)
	if err != nil {
		return nil, nil, err
	}
	z.Conn = conn
	z.Timeout = timeout
	z.ZkAddresses = zkAddresses
	return z, event, nil
}

func (z *zkClient) Close() {
	if z.Conn != nil {
		z.Conn.Close()
		z.Conn = nil
	}
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
				cs, ok := z.eventRegistry[e.Path]
				if ok {
					for _, c := range cs {
						*c <- struct{}{}
					}
				}
				z.Unlock()
			}
		}
	}
}

func (z *zkClient) CreateBasePath(basePath string) error {
	var temp string
	for _, subPath := range strings.Split(basePath, "/") {
		temp = path.Join(temp, "/", subPath)
		_, err := z.Conn.Create(temp, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				logger.Infof("zk.create(\"%s\") exists\n", temp)
			} else {
				logger.Errorf("zk.create(\"%s\") error(%v)\n", temp, perrors.WithStack(err))
				return perrors.WithMessagef(err, "zk.Create(path:%s)", basePath)
			}
		}
	}
	return nil
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
