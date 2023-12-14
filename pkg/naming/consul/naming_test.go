package consul

import (
	"X_IM/pkg/naming"
	"X_IM/pkg/x"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

const (
	namingAddr = "8.146.198.70"
	servAddr1  = "124.86.5.230"
	servAddr2  = "20.115.1.240"
)

func TestNaming(t *testing.T) {
	//consul http default port 8500
	nm, err := NewNaming(namingAddr + ":8500")
	assert.Nil(t, err)

	//0.init
	_ = nm.Deregister("test_1")
	_ = nm.Deregister("test_2")
	ok, err := nm.Find("test")
	assert.Nil(t, err)
	if assert.Equal(t, 0, len(ok)) {
		t.Log("init success")
	}

	serviceName := "test"
	//1.register test_1
	err = nm.Register(&naming.DefaultService{
		ID:        "test_1",
		Name:      serviceName,
		Namespace: "",
		Address:   servAddr1,
		Port:      8081,
		Protocol:  "ws",
		Tags:      []string{"tag1", "gateway"},
	})
	assert.Nil(t, err)

	// 2.find service
	// 注意这里find方法调用load方法
	// services := make([]x.ServiceRegistration, 0, len(catalogServices))
	// make时要防止初始化为nil
	// 会导致append时[0:nil,1:the_first_service]，len=2
	srvs, err := nm.Find(serviceName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(srvs))
	t.Log("find service:", srvs)

	var wg sync.WaitGroup
	wg.Add(1)

	//3.subscribe new service
	_ = nm.Subscribe(serviceName, func(services []x.ServiceRegistration) {
		t.Log(len(services))

		assert.Equal(t, 2, len(services))
		assert.Equal(t, "test_2", services[1].ServiceID())
		wg.Done()
	})
	time.Sleep(time.Second)

	//4.register test_2 to verify step3
	err = nm.Register(&naming.DefaultService{
		ID:        "test_2",
		Name:      serviceName,
		Namespace: "",
		Address:   servAddr2,
		Port:      8082,
		Protocol:  "ws",
		Tags:      []string{"tag2", "gateway"},
	})
	assert.Nil(t, err)
	//waiting for callback in Watch()
	wg.Wait()

	_ = nm.Unsubscribe(serviceName)

	//5.find service
	srvs, err = nm.Find(serviceName, "gateway")
	assert.Equal(t, 2, len(srvs))

	//6.find service to verify tag
	srvs, _ = nm.Find(serviceName, "tag2")
	assert.Equal(t, 1, len(srvs))
	assert.Equal(t, "test_2", srvs[0].ServiceID())

	//7.deregister test_2
	err = nm.Deregister("test_2")
	assert.Nil(t, err)

	//8.find service to verify deregister
	srvs, err = nm.Find(serviceName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(srvs))
	assert.Equal(t, "test_1", srvs[0].ServiceID())

	//9.deregister test_1
	err = nm.Deregister("test_1")
	assert.Nil(t, err)
}
