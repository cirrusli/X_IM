package X_IM

import (
	"X_IM/pkg"
	"X_IM/wire/pkt"
	"errors"
)

var ErrSessionNil = errors.New("err:session nil")

type SessionStorage interface {
	// Add a session
	Add(session *pkt.Session) error
	// Delete a session
	Delete(account string, channelId string) error
	// Get session by channelId
	Get(channelId string) (*pkt.Session, error)
	// GetLocations by accounts
	GetLocations(account ...string) ([]*pkg.Location, error)
	// GetLocation by account and device.
	// device parameter is not supported yet,there will be placed with ""
	GetLocation(account string, device string) (*pkg.Location, error)
}
