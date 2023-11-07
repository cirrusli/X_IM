package api

import (
	x "X_IM"
	"X_IM/naming"
	"X_IM/services/router/conf"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelectIDC(t *testing.T) {
	got := selectIDC("test1", &conf.Region{
		IDCs: []conf.IDC{
			{ID: "SH_ALI"},
			{ID: "HZ_ALI"},
			{ID: "SH_TENCENT"},
		},
		Slots: []byte{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2},
	})
	assert.NotNil(t, got)
	t.Log(got)
}
func TestSelectGateways(t *testing.T) {
	got := selectGateways("test11", []x.ServiceRegistration{
		&naming.DefaultService{ID: "g1"},
		&naming.DefaultService{ID: "g2"},
	}, 3)
	assert.Equal(t, len(got), 2)

	got = selectGateways("test11", []x.ServiceRegistration{
		&naming.DefaultService{ID: "g1"},
		&naming.DefaultService{ID: "g2"},
		&naming.DefaultService{ID: "g3"},
	}, 3)
	assert.Equal(t, len(got), 3)

	got = selectGateways("test11", []x.ServiceRegistration{
		&naming.DefaultService{ID: "g1"},
		&naming.DefaultService{ID: "g2"},
		&naming.DefaultService{ID: "g3"},
		&naming.DefaultService{ID: "g4"},
		&naming.DefaultService{ID: "g5"},
		&naming.DefaultService{ID: "g6"},
	}, 3)

	t.Log(got)
	assert.Equal(t, len(got), 3)
}
