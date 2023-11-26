package container

import (
	x "X_IM"
	"sync"
)

var l = log.WithField("object", "ClientsImpl")

// ClientMap 客户端集合
type ClientMap interface {
	Add(client x.Client)
	Remove(id string)
	Get(id string) (client x.Client, ok bool)
	Services(kvs ...string) []x.Service
}

type ClientsImpl struct {
	clients *sync.Map
}

func NewClients() ClientMap {
	return &ClientsImpl{
		clients: new(sync.Map),
	}
}

// Add Channel
func (ch *ClientsImpl) Add(client x.Client) {
	if client.ServiceID() == "" {
		l.WithField("func", "Add").Error("client id is required")
	}
	ch.clients.Store(client.ServiceID(), client)
}

// Remove added Channel
func (ch *ClientsImpl) Remove(id string) {
	ch.clients.Delete(id)
}

func (ch *ClientsImpl) Get(id string) (x.Client, bool) {
	if id == "" {
		l.WithField("func", "Get").Error("client id required")
	}

	val, ok := ch.clients.Load(id)
	if !ok {
		return nil, false
	}
	return val.(x.Client), true
}

// Services 返回服务列表，可以传一对
func (ch *ClientsImpl) Services(kvs ...string) []x.Service {
	kvLen := len(kvs)
	if kvLen != 0 && kvLen != 2 {
		return nil
	}
	arr := make([]x.Service, 0)
	ch.clients.Range(func(key, val any) bool {
		ser := val.(x.Service)
		if kvLen > 0 && ser.GetMeta()[kvs[0]] != kvs[1] {
			return true
		}
		arr = append(arr, ser)
		return true
	})
	return arr
}
