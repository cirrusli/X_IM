package storage

import (
	x "X_IM"
	"X_IM/pkg"
	"X_IM/pkg/wire/pkt"
	"fmt"
	"github.com/go-redis/redis/v7"
	"google.golang.org/protobuf/proto"
	"time"
)

const (
	LocationExpired = time.Hour * 48
)

type RedisStorage struct {
	cli *redis.Client
}

func NewRedisStorage(cli *redis.Client) x.SessionStorage {
	return &RedisStorage{cli: cli}
}
func (r *RedisStorage) Add(session *pkt.Session) error {
	// save x.Location
	loc := pkg.Location{
		ChannelID: session.ChannelID,
		GateID:    session.GateID,
	}
	locKey := KeyLocation(session.Account, "")
	err := r.cli.Set(locKey, loc.Bytes(), LocationExpired).Err()
	if err != nil {
		return err
	}
	// save session
	snKey := KeySession(session.ChannelID)
	buf, _ := proto.Marshal(session)
	err = r.cli.Set(snKey, buf, LocationExpired).Err()
	if err != nil {
		return err
	}
	return nil
}

// Delete a session
func (r *RedisStorage) Delete(account string, channelId string) error {
	locKey := KeyLocation(account, "")
	err := r.cli.Del(locKey).Err()
	if err != nil {
		return err
	}

	snKey := KeySession(channelId)
	err = r.cli.Del(snKey).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get session by sessionID
func (r *RedisStorage) Get(ChannelID string) (*pkt.Session, error) {
	snKey := KeySession(ChannelID)
	bts, err := r.cli.Get(snKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, x.ErrSessionNil
		}
		return nil, err
	}
	var session pkt.Session
	_ = proto.Unmarshal(bts, &session)
	return &session, nil
}

func (r *RedisStorage) GetLocations(accounts ...string) ([]*pkg.Location, error) {
	keys := KeyLocations(accounts...)
	list, err := r.cli.MGet(keys...).Result()
	if err != nil {
		return nil, err
	}
	var result = make([]*pkg.Location, 0)
	for _, l := range list {
		if l == nil {
			continue
		}
		var loc pkg.Location
		_ = loc.Unmarshal([]byte(l.(string)))
		result = append(result, &loc)
	}
	if len(result) == 0 {
		return nil, x.ErrSessionNil
	}
	return result, nil
}

func (r *RedisStorage) GetLocation(account string, device string) (*pkg.Location, error) {
	key := KeyLocation(account, device)
	bts, err := r.cli.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, x.ErrSessionNil
		}
		return nil, err
	}
	var loc pkg.Location
	_ = loc.Unmarshal(bts)
	return &loc, nil
}

// KeySession 转换为redis的key
func KeySession(channel string) string {
	return fmt.Sprintf("login:sn:%s", channel)
}

func KeyLocation(account, device string) string {
	if device == "" {
		return fmt.Sprintf("login:loc:%s", account)
	}
	return fmt.Sprintf("login:loc:%s:%s", account, device)
}

func KeyLocations(accounts ...string) []string {
	arr := make([]string, len(accounts))
	for i, account := range accounts {
		arr[i] = KeyLocation(account, "")
	}
	return arr
}

//func (r *RedisStorage) GetBySingleFlight(ChannelID string, g *singleflight.Group) (any, error) {
//	snKey := KeySession(ChannelID)
//	bts, err, _ := g.Do(snKey, func() (any, error) {
//
//		return
//	})
//
//	var session pkt.Session
//	_ = proto.Unmarshal(bts, &session)
//	return &session, nil
//}
