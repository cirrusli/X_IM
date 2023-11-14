package naming

import x "X_IM"

type Naming interface {
	// Find 服务发现，支持通过tag查询
	Find(serviceName string, tags ...string) ([]x.ServiceRegistration, error)
	Subscribe(serviceName string, callback func(services []x.ServiceRegistration)) error
	Unsubscribe(serviceName string) error
	Register(service x.ServiceRegistration) error
	Deregister(serviceID string) error
}
