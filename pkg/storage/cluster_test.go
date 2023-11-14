package storage

import (
	x "X_IM"
	"X_IM/pkg/wire/pkt"
	"github.com/gitstliu/go-redis-cluster"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedisClusterStorage_Add(t *testing.T) {
	cluster, err := redis.NewCluster(
		&redis.Options{
			StartNodes:   []string{"127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7002"},
			ConnTimeout:  50 * time.Millisecond,
			ReadTimeout:  50 * time.Millisecond,
			WriteTimeout: 50 * time.Millisecond,
			KeepAlive:    16,
			AliveTime:    60 * time.Second,
		})
	assert.Nil(t, err)
	cc := NewRedisClusterStorage(cluster)
	err = cc.Add(&pkt.Session{
		ChannelID: "ch1",
		GateID:    "gateway1",
		Account:   "test1",
		Device:    "Phone",
	})
	assert.Nil(t, err)

	_ = cc.Add(&pkt.Session{
		ChannelID: "ch2",
		GateID:    "gateway1",
		Account:   "test2",
		Device:    "Pc",
	})

	session, err := cc.Get("ch1")
	assert.Nil(t, err)
	t.Log(session)
	assert.Equal(t, "ch1", session.ChannelID)
	assert.Equal(t, "gateway1", session.GateID)
	assert.Equal(t, "test1", session.Account)

	arr, err := cc.GetLocations("test1", "test2")
	assert.Nil(t, err)
	t.Log(arr)
	loc := arr[1]

	arr, err = cc.GetLocations("test6")
	assert.Equal(t, x.ErrSessionNil, err)
	assert.Equal(t, 0, len(arr))

	assert.Equal(t, "ch2", loc.ChannelID)
	assert.Equal(t, "gateway1", loc.GateID)
}
