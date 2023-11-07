package naming

import (
	x "X_IM"
	"fmt"
)

// DefaultService Service Impl
type DefaultService struct {
	ID        string
	Name      string
	Address   string
	Port      int
	Protocol  string
	Namespace string
	Tags      []string
	Meta      map[string]string
}

func NewEntry(id, name, protocol string, address string, port int) x.ServiceRegistration {
	return &DefaultService{
		ID:       id,
		Name:     name,
		Address:  address,
		Port:     port,
		Protocol: protocol,
	}
}

// ServiceID returns the ServiceImpl ID
func (e *DefaultService) ServiceID() string { return e.ID }

func (e *DefaultService) GetNamespace() string { return e.Namespace }

func (e *DefaultService) ServiceName() string { return e.Name }

func (e *DefaultService) PublicAddress() string { return e.Address }

func (e *DefaultService) PublicPort() int { return e.Port }

func (e *DefaultService) GetProtocol() string { return e.Protocol }

func (e *DefaultService) DialURL() string {
	if e.Protocol == "tcp" {
		return fmt.Sprintf("%s:%d", e.Address, e.Port)
	}
	return fmt.Sprintf("%s://%s:%d", e.Protocol, e.Address, e.Port)
}

func (e *DefaultService) GetTags() []string { return e.Tags }

func (e *DefaultService) GetMeta() map[string]string { return e.Meta }

func (e *DefaultService) String() string {
	return fmt.Sprintf("ID:%s,Name:%s,Address:%s,Port:%d,Ns:%s,Tags:%v,Meta:%v",
		e.ID, e.Name, e.Address, e.Port, e.Namespace, e.Tags, e.Meta)
}
