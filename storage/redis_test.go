package storage

import (
	"X_IM"
	"X_IM/wire/pkt"
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

const (
	redisAddr = "8.146.198.70:6379"
)

func InitRedis(addr string, pass string) (*redis.Client, error) {
	redisDB := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	})

	_, err := redisDB.Ping().Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return redisDB, nil
}

func TestCRUD(t *testing.T) {
	password := "Redis:0617"
	c, err := InitRedis(redisAddr, password)
	assert.Nil(t, err)

	s := NewRedisStorage(c)
	err = s.Add(&pkt.Session{
		ChannelID: "channel1",
		GateID:    "gateway1",
		Account:   "account1",
		Device:    "device1",
	})
	assert.Nil(t, err)

	_ = s.Add(&pkt.Session{
		ChannelID: "channel2",
		GateID:    "gateway2",
		Account:   "account2",
		Device:    "device2",
	})

	session, err := s.Get("channel1")
	assert.Nil(t, err)
	t.Log(session)
	assert.Equal(t, "channel1", session.ChannelID)
	assert.Equal(t, "gateway1", session.GateID)
	assert.Equal(t, "account1", session.Account)
	assert.Equal(t, "device1", session.Device)

	arr, err := s.GetLocations("account1", "account2")
	assert.Nil(t, err)
	t.Log(arr)

	loc := arr[1]
	assert.Equal(t, "channel2", loc.ChannelID)

	arr, err = s.GetLocations("wrong_account")
	assert.Equal(t, X_IM.ErrSessionNil, err)
	assert.Equal(t, 0, len(arr))

}
