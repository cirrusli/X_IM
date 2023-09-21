package consul

import x "X_IM"

type Watch struct {
	Service   string
	Callback  func([]x.ServiceRegistration)
	WaitIndex uint64
	Quit      chan struct{}
}
