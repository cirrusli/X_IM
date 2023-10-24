package pkt

import (
	"X_IM/wire/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshal(t *testing.T) {
	bp := &BasicPkt{
		Code: CodePing,
	}

	bts := Marshal(bp)
	t.Log(bts)

	assert.Equal(t, common.MagicBasicPkt[1], bts[1])
	assert.Equal(t, common.MagicBasicPkt[2], bts[2])

	lp := New("login.signin")
	bts2 := Marshal(lp)
	t.Log(bts2)

	assert.Equal(t, common.MagicLogicPkt[1], bts2[1])
	assert.Equal(t, common.MagicLogicPkt[2], bts2[2])
}
