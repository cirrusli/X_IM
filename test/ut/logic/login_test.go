package logic

import (
	"X_IM/test/mock/dialer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const wsurl = "ws://localhost:8000"

// 需要先启动
func TestLogin(t *testing.T) {
	cli, err := dialer.Login(wsurl, "login_test_1")
	if cli == nil {
		t.Fatal("login failed, need run server first!")
	}
	defer cli.Close()
	assert.Nil(t, err)
	time.Sleep(time.Second * 2)

}
