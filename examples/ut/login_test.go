package ut

import (
	"X_IM/examples/dialer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const wsurl = "ws://localhost:8000"

func TestLogin(t *testing.T) {
	cli, err := dialer.Login(wsurl, "login_test_1")
	assert.Nil(t, err)
	time.Sleep(time.Second * 2)
	cli.Close()
}
