package serv

import (
	"X_IM/pkg/naming"
	"X_IM/pkg/wire/common"
	pkt2 "X_IM/pkg/wire/pkt"
	"X_IM/pkg/x"
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

	packet := pkt2.New(common.CommandChatUserTalk, pkt2.WithChannel(ksuid.New().String()))
	packet.AddStringMeta(MetaKeyApp, "x_im")
	packet.AddStringMeta(MetaKeyAccount, "test1")
	hit := rs.Lookup(&packet.Header, srvs)
	assert.Equal(t, "s6", hit)

	hits := make(map[string]int)
	for i := 0; i < 100; i++ {
		header := pkt2.Header{
			ChannelID: ksuid.New().String(),
			Meta: []*pkt2.Meta{
				{
					Type:  pkt2.MetaType_string,
					Key:   MetaKeyApp,
					Value: ksuid.New().String(),
				},
				{
					Type:  pkt2.MetaType_string,
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
