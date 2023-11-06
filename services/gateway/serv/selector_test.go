package serv

import (
	x "X_IM"
	"X_IM/naming"
	"X_IM/wire/common"
	"X_IM/wire/pkt"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelectorLookup(t *testing.T) {

	srvs := []x.Service{
		&naming.DefaultService{
			ID:   "s1",
			Meta: map[string]string{"zone": "zone_ali_01"},
		},
		&naming.DefaultService{
			ID:   "s2",
			Meta: map[string]string{"zone": "zone_ali_01"},
		},
		&naming.DefaultService{
			ID:   "s3",
			Meta: map[string]string{"zone": "zone_ali_01"},
		},
		&naming.DefaultService{
			ID:   "s4",
			Meta: map[string]string{"zone": "zone_ali_02"},
		},
		&naming.DefaultService{
			ID:   "s5",
			Meta: map[string]string{"zone": "zone_ali_02"},
		},
		&naming.DefaultService{
			ID:   "s6",
			Meta: map[string]string{"zone": "zone_ali_03"},
		},
	}

	rs, err := NewRouteSelector("../route.json")
	assert.Nil(t, err)

	packet := pkt.New(common.CommandChatUserTalk, pkt.WithChannel(ksuid.New().String()))
	packet.AddStringMeta(MetaKeyApp, "kim")
	packet.AddStringMeta(MetaKeyAccount, "test1")
	hit := rs.Lookup(&packet.Header, srvs)
	assert.Equal(t, "s6", hit)

	hits := make(map[string]int)
	for i := 0; i < 100; i++ {
		header := pkt.Header{
			ChannelID: ksuid.New().String(),
			Meta: []*pkt.Meta{
				{
					Type:  pkt.MetaType_string,
					Key:   MetaKeyApp,
					Value: ksuid.New().String(),
				},
				{
					Type:  pkt.MetaType_string,
					Key:   MetaKeyAccount,
					Value: ksuid.New().String(),
				},
			},
		}
		hit = rs.Lookup(&header, srvs)
		hits[hit]++
	}
	t.Log(hits)
}
